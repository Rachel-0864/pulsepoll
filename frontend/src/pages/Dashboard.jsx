import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  Plus,
  ArrowUpRight,
  Radio,
  BarChart3,
  Vote,
  Activity,
  Sparkles,
  RefreshCw,
} from "lucide-react";
import { api } from "../api/api";

export default function Dashboard() {
  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadPolls = async () => {
    try {
      setLoading(true);
      setError("");

      const response = await api.getMyPolls();

      // Supports either:
      // { polls: [...] }
      // or directly [...]
      const data = response.data;

      const receivedPolls = Array.isArray(data)
        ? data
        : data.polls || [];

      setPolls(receivedPolls);
    } catch (err) {
      console.error("Failed to load polls:", err);

      const message =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Could not load your polls.";

      setError(message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPolls();
  }, []);

  const totalVotes = polls.reduce((total, poll) => {
    return (
      total +
      (poll.options || []).reduce(
        (sum, option) => sum + Number(option.count || 0),
        0
      )
    );
  }, 0);

  const livePolls = polls.filter(
    (poll) => poll.isActive !== false
  ).length;

  return (
    <div className="page dashboard-page">

      {/* HERO / HEADER */}
      <section className="dashboard-hero">
        <div className="dashboard-heading">
          <div className="dashboard-title">
            <div className="eyebrow">
              <span className="eyebrow-dot" />
              YOUR WORKSPACE
            </div>

            <h1>
              Turn questions into
              <span className="gradient-text"> live conversations.</span>
            </h1>

            <p>
              Create polls, share them with your audience and watch
              responses arrive in real time.
            </p>
          </div>

          <Link
            className="button primary dashboard-create-button"
            to="/create"
          >
            <Plus size={18} />
            New Poll
          </Link>
        </div>

        <div className="hero-glow glow-one" />
        <div className="hero-glow glow-two" />
      </section>

      {/* STATS */}
      <section className="stats-grid">

        <div className="stat-card card">
          <div className="stat-icon">
            <BarChart3 size={20} />
          </div>

          <div className="stat-content">
            <span>Total polls</span>
            <strong>{loading ? "—" : polls.length}</strong>
          </div>

          <div className="stat-decoration">
            <BarChart3 size={55} />
          </div>
        </div>

        <div className="stat-card card">
          <div className="stat-icon live-stat-icon">
            <Radio size={20} />
          </div>

          <div className="stat-content">
            <span>Live polls</span>
            <strong>{loading ? "—" : livePolls}</strong>
          </div>

          <div className="live-indicator">
            <i />
            LIVE
          </div>
        </div>

        <div className="stat-card card">
          <div className="stat-icon vote-stat-icon">
            <Vote size={20} />
          </div>

          <div className="stat-content">
            <span>Total responses</span>
            <strong>{loading ? "—" : totalVotes}</strong>
          </div>

          <div className="stat-decoration">
            <Activity size={55} />
          </div>
        </div>

      </section>

      {/* SECTION HEADER */}
      <section className="polls-section">

        <div className="section-header">
          <div>
            <div className="section-title-row">
              <h2>Your polls</h2>

              {!loading && polls.length > 0 && (
                <span className="count-badge">
                  {polls.length}
                </span>
              )}
            </div>

            <p>
              Monitor your polls and see how your audience is responding.
            </p>
          </div>

          <button
            className="refresh-button"
            onClick={loadPolls}
            disabled={loading}
            title="Refresh polls"
          >
            <RefreshCw
              size={17}
              className={loading ? "spin" : ""}
            />
            Refresh
          </button>
        </div>

        {/* ERROR */}
        {error && (
          <div className="dashboard-alert">
            <div>
              <strong>Unable to load polls</strong>
              <span>{error}</span>
            </div>

            <button onClick={loadPolls}>
              Try again
            </button>
          </div>
        )}

        {/* LOADING */}
        {loading && (
          <div className="poll-grid">
            {[1, 2, 3].map((item) => (
              <div
                className="poll-card skeleton-card card"
                key={item}
              >
                <div className="skeleton skeleton-small" />
                <div className="skeleton skeleton-title" />
                <div className="skeleton skeleton-line" />
                <div className="skeleton skeleton-bars" />
              </div>
            ))}
          </div>
        )}

        {/* EMPTY STATE */}
        {!loading && !error && polls.length === 0 && (
          <div className="empty-polls card">
            <div className="empty-icon">
              <Sparkles size={28} />
            </div>

            <h3>Your first poll starts here</h3>

            <p>
              Ask a question, add your options and share the poll
              with your audience. Results update live as people vote.
            </p>

            <Link
              className="button primary"
              to="/create"
            >
              <Plus size={18} />
              Create your first poll
            </Link>
          </div>
        )}

        {/* POLLS */}
        {!loading && !error && polls.length > 0 && (
          <div className="poll-grid">

            {polls.map((poll) => {
              const options = poll.options || [];

              const total = options.reduce(
                (sum, option) =>
                  sum + Number(option.count || 0),
                0
              );

              const isLive = poll.isActive !== false;

              return (
                <Link
                  to={`/poll/${poll.id}`}
                  className="poll-card card"
                  key={poll.id}
                >

                  {/* CARD TOP */}
                  <div className="poll-card-top">

                    <span
                      className={`status ${
                        isLive ? "status-live" : "status-closed"
                      }`}
                    >
                      <i />
                      {isLive ? "Live" : "Closed"}
                    </span>

                    <span className="poll-open">
                      <ArrowUpRight size={18} />
                    </span>
                  </div>

                  {/* QUESTION */}
                  <h3>{poll.question}</h3>

                  {/* META */}
                  <div className="poll-card-meta">
                    <span>
                      {options.length}{" "}
                      {options.length === 1
                        ? "option"
                        : "options"}
                    </span>

                    <span>
                      {total}{" "}
                      {total === 1
                        ? "response"
                        : "responses"}
                    </span>
                  </div>

                  {/* VOTE DISTRIBUTION */}
                  <div className="poll-visual">

                    {options.length > 0 ? (
                      options.map((option) => {
                        const count = Number(
                          option.count || 0
                        );

                        const percentage =
                          total > 0
                            ? (count / total) * 100
                            : 0;

                        return (
                          <div
                            className="poll-option-bar"
                            key={option.id}
                          >
                            <div
                              className="poll-option-fill"
                              style={{
                                width: `${Math.max(
                                  total > 0
                                    ? percentage
                                    : 4,
                                  4
                                )}%`,
                              }}
                            />
                          </div>
                        );
                      })
                    ) : (
                      <div className="no-options">
                        No options yet
                      </div>
                    )}

                  </div>

                  {/* FOOTER */}
                  <div className="poll-card-footer">

                    <span>
                      <Activity size={14} />
                      Live results
                    </span>

                    <span className="view-poll">
                      View poll
                      <ArrowUpRight size={14} />
                    </span>

                  </div>

                </Link>
              );
            })}

          </div>
        )}

      </section>
    </div>
  );
}