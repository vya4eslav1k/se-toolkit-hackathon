import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import PollFeed from './components/PollFeed';
import CreatePoll from './components/CreatePoll';
import ViewPoll from './components/ViewPoll';
import Results from './components/Results';
import ThemeToggle from './components/ThemeToggle';
import './App.css';

const App: React.FC = () => {
  return (
    <Router>
      <div className="app">
        <nav className="navbar">
          <div className="container">
            <Link to="/" className="navbar-brand">
              📊 Poll Service
            </Link>
            <div className="navbar-links">
              <Link to="/">Feed</Link>
              <Link to="/create">Create Poll</Link>
              <ThemeToggle />
            </div>
          </div>
        </nav>
        <main>
          <Routes>
            <Route path="/" element={<PollFeed />} />
            <Route path="/create" element={<CreatePoll />} />
            <Route path="/poll/:id" element={<ViewPoll />} />
            <Route path="/results/:id" element={<Results />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App;
