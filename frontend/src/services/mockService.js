import { initialPolls } from "../data/mockPolls";

const KEY = "pulsepoll_mock_polls";

function read() {
  try {
    const stored = JSON.parse(localStorage.getItem(KEY));
    return stored?.length ? stored : initialPolls;
  } catch {
    return initialPolls;
  }
}
function write(polls) { localStorage.setItem(KEY, JSON.stringify(polls)); }

export const mockService = {
  list() { return Promise.resolve(read()); },
  get(id) { return Promise.resolve(read().find(p => p.id === id)); },
  create(question, options) {
    const poll = {
      id: `${Date.now()}`,
      question,
      options: options.map((text, i) => ({ id: `option-${i + 1}`, text, count: 0 })),
      creatorId: "demo-user",
      isActive: true,
      createdAt: new Date().toISOString()
    };
    const polls = [poll, ...read()];
    write(polls);
    return Promise.resolve(poll);
  },
  vote(id, optionId) {
    const polls = read();
    const poll = polls.find(p => p.id === id);
    if (!poll) throw new Error("Poll not found.");
    const option = poll.options.find(o => o.id === optionId);
    if (!option) throw new Error("Invalid option.");
    option.count += 1;
    write(polls);
    return Promise.resolve(poll);
  }
};