import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  Copy,
  Check,
  Radio,
  RefreshCw,
  Users,
} from "lucide-react";

import VoteButtons from "../components/VoteButtons";
import PollResultsChart from "../components/PollResultsChart";
import useWebSocket from "../hooks/useWebSocket";
import { api } from "../api/api";

function normalizePoll(data) {
  const poll = data?.poll || data;

  if (!poll) {
    return null;
  }

  return {
    ...poll,
    options: (poll.options || []).map((option) => ({
      ...option,
      count: Number(
        data?.counts?.[option.id] ??
          option.count ??
          0
      ),
    })),
  };
}

export default function PollView() {
  const { id } = useParams();

  const [poll, setPoll] = useState(null);
  const [selected, setSelected] = useState("");
  const [voted, setVoted] = useState(false);
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(true);
  const [voting, setVoting] = useState(false);
  const [error, setError] = useState("");

  // =========================================================
  // LOAD POLL
  // =========================================================

  const loadPoll = useCallback(async () => {
    if (!id) return;

    try {
      setLoading(true);
      setError("");

      const response = await api.getPoll(id);

      const normalized = normalizePoll(response.data);

      if (!normalized) {
        throw new Error("Poll data was not returned by the server.");
      }

      setPoll(normalized);
    } catch (err) {
      console.error("Failed to load poll:", err);

      const message =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Could not load this poll.";

      setError(message);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadPoll();
  }, [loadPoll]);

  // =========================================================
  // WEBSOCKET LIVE UPDATES
  // =========================================================

  const handleUpdate = useCallback((update) => {
    if (!update?.counts) {
      return;
    }

    setPoll((previousPoll) => {
      if (!previousPoll) {
        return previousPoll;
      }

      return {
        ...previousPoll,

        options: previousPoll.options.map((option) => ({
          ...option,
          count: Number(
            update.counts[option.id] ??
              option.count ??
              0
          ),
        })),
      };
    });
  }, []);

  const { status } = useWebSocket(id, handleUpdate);

  // =========================================================
  // VOTE
  // =========================================================

  const vote = async () => {
    if (!selected || voted || voting) {
      return;
    }

    try {
      setError("");
      setVoting(true);

      const response = await api.vote(id, selected);

      /*
       * IMPORTANT:
       *
       * Backend returns:
       *
       * {
       *   counts: {
       *     opt1: "1",
       *     opt2: "0"
       *   }
       * }
       *
       * It does NOT return a complete poll.
       *
       * Therefore we merge the returned counts into the
       * existing poll instead of calling normalizePoll()
       * on the vote response.
       */

      const counts = response?.data?.counts;

      if (counts) {
        setPoll((previousPoll) => {
          if (!previousPoll) {
            return previousPoll;
          }

          return {
            ...previousPoll,

            options: previousPoll.options.map(
              (option) => ({
                ...option,
                count: Number(
                  counts[option.id] ??
                    option.count ??
                    0
                ),
              })
            ),
          };
        });
      }

      setVoted(true);
      setSelected("");
    } catch (err) {
      console.error("Vote error:", err);

      const message =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Could not record your vote.";

      setError(message);
    } finally {
      setVoting(false);
    }
  };

  // =========================================================
  // COPY LINK
  // =========================================================

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(
        window.location.href
      );
    } catch (err) {
      console.error("Could not copy link:", err);
    }

    setCopied(true);

    setTimeout(() => {
      setCopied(false);
    }, 1500);
  };

  // =========================================================
  // LOADING
  // =========================================================

  if (loading) {
    return (
      <div className="poll-page">
        <div className="poll-loading card">
          <RefreshCw
            className="spin"
            size={24}
          />

          <span>
            Loading live poll...
          </span>
        </div>
      </div>
    );
  }

  // =========================================================
  // ERROR
  // =========================================================

  if (error && !poll) {
    return (
      <div className="poll-page">
        <div className="poll-error card">
          <div className="poll-error-icon">
            <Radio size={25} />
          </div>

          <h2>
            Unable to load this poll
          </h2>

          <p>{error}</p>

          <button
            className="button primary"
            onClick={loadPoll}
          >
            <RefreshCw size={17} />
            Try again
          </button>
        </div>

        <Link
          className="back-link"
          to="/"
        >
          ← Back to home
        </Link>
      </div>
    );
  }

  if (!poll) {
    return null;
  }

  // =========================================================
  // TOTAL VOTES
  // =========================================================

  const totalVotes = poll.options.reduce(
    (total, option) =>
      total + Number(option.count || 0),
    0
  );

  // =========================================================
  // CONNECTION STATUS
  // =========================================================

  const connectionLabel =
    status === "connected"
      ? "Live connection"
      : status === "connecting"
      ? "Connecting..."
      : status === "error"
      ? "Connection error"
      : "Offline";

  // =========================================================
  // UI
  // =========================================================

  return (
    <div className="poll-page professional-poll-page">

      <div className="poll-layout">

        {/* =================================================
            VOTING CARD
        ================================================= */}

        <section className="card vote-card professional-vote-card">

          <div className="live-header">

            <span className="status status-live">
              <i />
              Live poll
            </span>

            <span
              className={`connection connection-${status}`}
            >
              <Radio size={14} />

              {connectionLabel}
            </span>

          </div>

          <div className="poll-question-area">

            <span className="poll-question-label">
              QUESTION
            </span>

            <h1>
              {poll.question}
            </h1>

            <p className="vote-subtitle">
              Choose one option below to cast
              your vote.
            </p>

          </div>

          {error && (
            <div className="alert">
              {error}
            </div>
          )}

          <VoteButtons
            options={poll.options}
            selected={selected}
            setSelected={setSelected}
            onVote={vote}
            disabled={voted || voting}
          />

          {voted && (
            <div className="success">
              <Check size={16} />

              <span>
                Your vote has been recorded.
              </span>
            </div>
          )}

          <div className="vote-card-footer">

            <div className="audience-count">
              <Users size={15} />

              <span>
                {totalVotes}{" "}
                {totalVotes === 1
                  ? "response"
                  : "responses"}
              </span>
            </div>

            <button
              className="share-button"
              onClick={copyLink}
              type="button"
            >
              {copied ? (
                <Check size={16} />
              ) : (
                <Copy size={16} />
              )}

              {copied
                ? "Copied"
                : "Copy poll link"}
            </button>

          </div>

        </section>

        {/* =================================================
            RESULTS CARD
        ================================================= */}

        <section className="card results-card professional-results-card">

          <div className="results-heading">

            <div>

              <span className="eyebrow">
                LIVE ANALYTICS
              </span>

              <h2>
                Results
              </h2>

              <p className="results-description">
                Responses update automatically
                as people vote.
              </p>

            </div>

            <span className="live-pill">
              <i />
              LIVE
            </span>

          </div>

          <div className="results-total">

            <strong>
              {totalVotes}
            </strong>

            <span>
              total{" "}
              {totalVotes === 1
                ? "vote"
                : "votes"}
            </span>

          </div>

          <PollResultsChart
            options={poll.options}
          />

        </section>

      </div>

      <Link
        className="back-link"
        to="/"
      >
        ← Back to home
      </Link>

    </div>
  );
}