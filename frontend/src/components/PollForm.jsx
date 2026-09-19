import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";

export default function PollForm({ onSubmit, loading }) {
  const [question, setQuestion] = useState("");
  const [options, setOptions] = useState(["", ""]);

  const updateOption = (i, value) => {
    setOptions(prev => prev.map((x, idx) => idx === i ? value : x));
  };

  const addOption = () => setOptions(prev => [...prev, ""]);
  const removeOption = (i) => {
    if (options.length <= 2) return;
    setOptions(prev => prev.filter((_, idx) => idx !== i));
  };

  const submit = (e) => {
    e.preventDefault();
    const cleaned = options.map(x => x.trim()).filter(Boolean);
    if (!question.trim()) return;
    if (cleaned.length < 2) return;
    onSubmit({ question: question.trim(), options: cleaned });
  };

  return (
    <form className="card form-card" onSubmit={submit}>
      <label>Poll question</label>
      <input
        className="input question-input"
        value={question}
        onChange={e => setQuestion(e.target.value)}
        placeholder="What would you like to ask?"
        maxLength={180}
        required
      />

      <div className="form-section-title">
        <label>Answer options</label>
        <span>{options.length} options</span>
      </div>

      <div className="options-editor">
        {options.map((option, i) => (
          <div className="option-edit-row" key={i}>
            <span className="option-number">{i + 1}</span>
            <input
              className="input"
              value={option}
              onChange={e => updateOption(i, e.target.value)}
              placeholder={`Option ${i + 1}`}
              maxLength={100}
              required
            />
            <button type="button" className="icon-button" onClick={() => removeOption(i)} disabled={options.length <= 2}>
              <Trash2 size={17} />
            </button>
          </div>
        ))}
      </div>

      <button type="button" className="add-option" onClick={addOption}>
        <Plus size={17} /> Add option
      </button>

      <button className="button primary full" disabled={loading}>
        {loading ? "Creating..." : "Create Poll"}
      </button>
    </form>
  );
}