package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

type UserState struct {
	State        string
	SelectedPoll *PollItem
}

type PollItem struct {
	ID    string
	Title string
	Index int
}

type Bot struct {
	b      *tele.Bot
	apiURL string
	users  map[int64]*UserState
}

func New(token, apiURL string) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second, AllowedUpdates: []string{"message", "callback_query"}},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot := &Bot{
		b:      b,
		apiURL: apiURL,
		users:  make(map[int64]*UserState),
	}

	bot.setupHandlers()

	return bot, nil
}

func (b *Bot) Start() {
	fmt.Println("Telegram bot started")
	b.b.Start()
}

func (b *Bot) replyMarkup() *tele.ReplyMarkup {
	kb := &tele.ReplyMarkup{
		ResizeKeyboard: true,
	}
	kb.Reply(
		kb.Row(kb.Text("📊 Лента")),
	)
	return kb
}

func (b *Bot) getState(id int64) *UserState {
	if state, ok := b.users[id]; ok {
		return state
	}
	state := &UserState{State: "idle"}
	b.users[id] = state
	return state
}

func (b *Bot) clearState(id int64) {
	delete(b.users, id)
}

func (b *Bot) apiGet(path string, telegramID int64) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", b.apiURL+path, nil)
	req.AddCookie(&http.Cookie{
		Name:  "voter_id",
		Value: fmt.Sprintf("tg-%d", telegramID),
	})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (b *Bot) apiPost(path string, data interface{}, telegramID int64) ([]byte, error) {
	body, _ := json.Marshal(data)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", b.apiURL+path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "voter_id",
		Value: fmt.Sprintf("tg-%d", telegramID),
	})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (b *Bot) setupHandlers() {
	b.b.Handle("/start", func(c tele.Context) error {
		b.clearState(c.Sender().ID)
		return c.Send("👋 Привет! Я бот для опросов.\n\n" +
			"/feed — лента опросов\n" +
			"/results <id> — результаты опроса")
	})

	b.b.Handle("/feed", func(c tele.Context) error {
		b.clearState(c.Sender().ID)
		return b.showFeed(c)
	})

	b.b.Handle("📊 Лента", func(c tele.Context) error {
		b.clearState(c.Sender().ID)
		return b.showFeed(c)
	})

	b.b.Handle("/results", func(c tele.Context) error {
		b.clearState(c.Sender().ID)
		args := strings.Split(c.Text(), " ")
		if len(args) < 2 {
			return c.Send("❌ Использование: /results <id опроса>")
		}
		pollID := args[1]
		return b.showResults(c, pollID)
	})

	b.b.Handle(tele.OnText, func(c tele.Context) error {
		state := b.getState(c.Sender().ID)

		switch state.State {
		case "chose_feed_poll":
			return b.handleFeedSelection(c, state, c.Text())
		case "chose_vote_option":
			return b.handleVote(c, state, c.Text())
		default:
			return c.Send("Нажми /feed или /results <id>")
		}
	})

	// Callback button handlers
	b.b.Handle(&tele.Btn{Unique: "results"}, func(c tele.Context) error {
		payload := c.Data()
		fmt.Printf("[btn] results payload=%s\n", payload)
		if payload == "latest" || payload == "" {
			return c.Respond(&tele.CallbackResponse{Text: "Выбери опрос из ленты"})
		}
		b.clearState(c.Sender().ID)
		c.Respond(&tele.CallbackResponse{Text: ""})
		return b.showResults(c, payload)
	})

	b.b.Handle(&tele.Btn{Unique: "feed_back"}, func(c tele.Context) error {
		fmt.Println("[btn] feed_back")
		b.clearState(c.Sender().ID)
		c.Respond(&tele.CallbackResponse{Text: ""})
		return b.showFeed(c)
	})
}

func (b *Bot) showFeed(c tele.Context) error {
	data, err := b.apiGet("/polls", c.Sender().ID)
	if err != nil {
		return c.Send("❌ Ошибка загрузки опросов")
	}

	var polls []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	if err := json.Unmarshal(data, &polls); err != nil {
		return c.Send("❌ Ошибка обработки данных")
	}

	if len(polls) == 0 {
		return c.Send("📭 Лента пуста. Создай опрос через веб-интерфейс.")
	}

	msg := "📊 <b>Лента опросов</b>\n\n"
	for i, p := range polls {
		msg += fmt.Sprintf("<b>%d</b> — %s\n", i+1, p.Title)
	}
	msg += "\nВведи номер опроса:"

	state := b.getState(c.Sender().ID)
	state.State = "chose_feed_poll"
	state.SelectedPoll = nil

	return c.Send(msg, b.replyMarkup(), tele.ModeHTML)
}

func (b *Bot) handleFeedSelection(c tele.Context, state *UserState, text string) error {
	num, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || num < 1 {
		return c.Send("❌ Введи корректный номер")
	}

	// Re-fetch polls to get the selected one
	data, err := b.apiGet("/polls", c.Sender().ID)
	if err != nil {
		return c.Send("❌ Ошибка загрузки")
	}

	var polls []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	json.Unmarshal(data, &polls)

	if num > len(polls) {
		return c.Send(fmt.Sprintf("❌ Номер должен быть от 1 до %d", len(polls)))
	}

	selected := polls[num-1]
	return b.showPoll(c, selected.ID)
}

func (b *Bot) showPoll(c tele.Context, pollID string) error {
	data, err := b.apiGet("/polls/"+pollID, c.Sender().ID)
	if err != nil {
		return c.Send("❌ Опрос не найден")
	}

	var resp struct {
		Poll struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Questions []struct {
				ID      string `json:"id"`
				Text    string `json:"text"`
				Options []struct {
					ID   string `json:"id"`
					Text string `json:"text"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"poll"`
		VoteStatus map[string]bool `json:"vote_status"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return c.Send("❌ Ошибка обработки данных")
	}

	// Check if already voted
	allVoted := true
	for _, voted := range resp.VoteStatus {
		if !voted {
			allVoted = false
			break
		}
	}

	if allVoted && len(resp.VoteStatus) > 0 {
		msg := fmt.Sprintf("✅ Ты уже проголосовал в опросе \"%s\"", resp.Poll.Title)

		kb := &tele.ReplyMarkup{}
		kb.Inline(kb.Row(
			kb.Data("📊 Результаты", "results", resp.Poll.ID),
		))
		return c.Send(msg, kb)
	}

	// Show first unanswered question
	for _, q := range resp.Poll.Questions {
		if resp.VoteStatus[q.ID] {
			continue
		}

		msg := fmt.Sprintf("📝 <b>%s</b>\n\n<b>%s</b>\n", resp.Poll.Title, q.Text)
		for i, opt := range q.Options {
			msg += fmt.Sprintf("<b>%d</b> — %s\n", i+1, opt.Text)
		}
		msg += "\nВведи номер варианта:"

		kb := &tele.ReplyMarkup{}
		kb.Inline(kb.Row(
			kb.Data("📊 Результаты", "results", resp.Poll.ID),
		))

		state := b.getState(c.Sender().ID)
		state.State = "chose_vote_option"
		state.SelectedPoll = &PollItem{ID: resp.Poll.ID}

		return c.Send(msg, kb, tele.ModeHTML)
	}

	kb := &tele.ReplyMarkup{}
	kb.Inline(kb.Row(
		kb.Data("📊 Результаты", "results", resp.Poll.ID),
		kb.Data("🔙 К ленте", "feed_back"),
	))
	return c.Send("✅ Все вопросы отвечены!", kb)
}

func (b *Bot) handleVote(c tele.Context, state *UserState, text string) error {
	num, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || num < 1 {
		return c.Send("❌ Введи корректный номер")
	}

	// Get poll data
	pollID := state.SelectedPoll.ID
	if pollID == "" {
		b.clearState(c.Sender().ID)
		return c.Send("❌ Сессия истекла. Начни заново: /feed")
	}

	data, err := b.apiGet("/polls/"+pollID, c.Sender().ID)
	if err != nil {
		return c.Send("❌ Ошибка загрузки")
	}

	var resp struct {
		Poll struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Questions []struct {
				ID      string `json:"id"`
				Options []struct {
					ID   string `json:"id"`
					Text string `json:"text"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"poll"`
		VoteStatus map[string]bool `json:"vote_status"`
	}

	json.Unmarshal(data, &resp)

	// Find first unanswered question
	var targetQuestion *struct {
		ID      string `json:"id"`
		Options []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"options"`
	}

	for i := range resp.Poll.Questions {
		if !resp.VoteStatus[resp.Poll.Questions[i].ID] {
			targetQuestion = &resp.Poll.Questions[i]
			break
		}
	}

	if targetQuestion == nil {
		return c.Send("✅ Все вопросы отвечены!")
	}

	if num > len(targetQuestion.Options) {
		return c.Send(fmt.Sprintf("❌ Номер от 1 до %d", len(targetQuestion.Options)))
	}

	selectedOption := targetQuestion.Options[num-1]

	// Submit vote
	voteData := map[string]interface{}{
		"answers": []map[string]string{
			{
				"question_id": targetQuestion.ID,
				"option_id":   selectedOption.ID,
			},
		},
	}

	_, err = b.apiPost("/polls/"+pollID+"/vote", voteData, c.Sender().ID)
	if err != nil {
		return c.Send("❌ Ошибка голосования. Возможно ты уже голосовал.")
	}

	// Re-fetch poll to check for more questions
	data2, err := b.apiGet("/polls/"+pollID, c.Sender().ID)
	if err != nil {
		return c.Send("❌ Ошибка загрузки")
	}

	var resp2 struct {
		Poll struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Questions []struct {
				ID      string `json:"id"`
				Text    string `json:"text"`
				Options []struct {
					ID   string `json:"id"`
					Text string `json:"text"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"poll"`
		VoteStatus map[string]bool `json:"vote_status"`
	}

	json.Unmarshal(data2, &resp2)

	// Check for more unanswered questions
	for _, q := range resp2.Poll.Questions {
		if resp2.VoteStatus[q.ID] {
			continue
		}

		// Show next question
		msg := fmt.Sprintf("📝 <b>%s</b>\n\n<b>%s</b>\n", resp2.Poll.Title, q.Text)
		for i, opt := range q.Options {
			msg += fmt.Sprintf("<b>%d</b> — %s\n", i+1, opt.Text)
		}
		msg += "\nВведи номер варианта:"

		kb := &tele.ReplyMarkup{}
		kb.Inline(kb.Row(
			kb.Data("📊 Результаты", "results", pollID),
		))

		state.SelectedPoll = &PollItem{ID: pollID}
		state.State = "chose_vote_option"

		return c.Send(msg, kb, tele.ModeHTML)
	}

	// All questions answered
	msg := "✅ Все вопросы отвечены!"

	kb := &tele.ReplyMarkup{}
	kb.Inline(kb.Row(
		kb.Data("📊 Результаты", "results", pollID),
		kb.Data("🔙 К ленте", "feed_back"),
	))
	b.clearState(c.Sender().ID)
	return c.Send(msg, kb)
}

func (b *Bot) showResults(c tele.Context, pollID string) error {
	data, err := b.apiGet("/polls/"+pollID+"/results", c.Sender().ID)
	if err != nil {
		return c.Send("❌ Не удалось загрузить результаты. Возможно у тебя нет доступа.")
	}

	var resp struct {
		Poll struct {
			Title     string `json:"title"`
			Questions []struct {
				Text    string `json:"text"`
				Options []struct {
					Text  string `json:"text"`
					Votes int    `json:"votes"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"poll"`
		TotalVotes int  `json:"total_votes"`
		CanView    bool `json:"can_view"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return c.Send("❌ Ошибка обработки данных")
	}

	if !resp.CanView {
		return c.Send("🔒 Результаты недоступны. Проголосуй первым или это только для автора.")
	}

	msg := fmt.Sprintf("📊 <b>%s</b>\nГолосов: %d\n\n", resp.Poll.Title, resp.TotalVotes)

	for _, q := range resp.Poll.Questions {
		msg += fmt.Sprintf("<b>%s</b>\n", q.Text)
		for _, opt := range q.Options {
			pct := 0
			if resp.TotalVotes > 0 {
				pct = opt.Votes * 100 / resp.TotalVotes
			}
			bar := ""
			for i := 0; i < pct/5; i++ {
				bar += "█"
			}
			msg += fmt.Sprintf("  %s — %d (%d%%) %s\n", opt.Text, opt.Votes, pct, bar)
		}
		msg += "\n"
	}

	kb := &tele.ReplyMarkup{}
	kb.Inline(kb.Row(
		kb.Data("🔙 К ленте", "feed_back"),
	))

	return c.Send(msg, kb, tele.ModeHTML)
}
