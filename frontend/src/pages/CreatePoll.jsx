import { useState } from "react";
import { useNavigate } from "react-router-dom";
import PollForm from "../components/PollForm";
import { api } from "../api/api";

export default function CreatePoll() {
  const navigate = useNavigate();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const create = async ({ question, options }) => {
    setError("");
    setLoading(true);

    try {
      const response = await api.createPoll({
        question,
        options,
      });

      const data = response.data;

      // Backend may return:
      // { poll: {...} }
      // or directly: {...}
      const poll = data.poll || data;

      if (!poll?.id) {
        throw new Error("Poll was created, but no poll ID was returned.");
      }

      navigate(`/poll/${poll.id}`);
    } catch (err) {
      console.error("Create poll error:", err);

      const message =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Could not create the poll. Please try again.";

      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page narrow create-poll-page">
      <div className="page-heading create-poll-heading">
        <span className="eyebrow">
          <span className="eyebrow-dot" />
          NEW POLL
        </span>

        <h1>Create a poll</h1>

        <p>
          Ask a question, add your options, and let your audience
          vote in real time.
        </p>
      </div>

      {error && (
        <div className="dashboard-alert create-poll-alert">
          <div>
            <strong>Could not create poll</strong>
            <span>{error}</span>
          </div>

          <button onClick={() => setError("")}>
            Dismiss
          </button>
        </div>
      )}

      <PollForm
        onSubmit={create}
        loading={loading}
      />
    </div>
  );
}