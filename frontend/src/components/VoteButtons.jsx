export default function VoteButtons({ options, selected, setSelected, onVote, disabled }) {
  return (
    <div className="vote-options">
      {options.map((option) => (
        <button
          type="button"
          key={option.id}
          className={`vote-option ${selected === option.id ? "selected" : ""}`}
          onClick={() => setSelected(option.id)}
          disabled={disabled}
        >
          <span className="radio-dot">{selected === option.id && <span />}</span>
          <span>{option.text}</span>
        </button>
      ))}
      <button className="button primary full" disabled={!selected || disabled} onClick={onVote}>
        {disabled ? "Vote submitted" : "Submit Vote"}
      </button>
    </div>
  );
}