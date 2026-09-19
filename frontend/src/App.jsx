import { Routes, Route } from "react-router-dom";
import Navbar from "./components/Navbar";
import AuthGuard from "./components/AuthGuard";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Signup from "./pages/Signup";
import Dashboard from "./pages/Dashboard";
import CreatePoll from "./pages/CreatePoll";
import PollView from "./pages/PollView";
import NotFound from "./pages/NotFound";

export default function App() {
  return (
    <div className="app-shell">
      <Navbar />
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          <Route path="/dashboard" element={<AuthGuard><Dashboard /></AuthGuard>} />
          <Route path="/create" element={<AuthGuard><CreatePoll /></AuthGuard>} />
          <Route path="/poll/:id" element={<PollView />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </main>
      <footer className="footer">PulsePoll · Live voting without page refresh</footer>
    </div>
  );
}