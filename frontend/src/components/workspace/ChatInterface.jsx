import { useEffect, useMemo, useState } from "react";
import PropTypes from "prop-types";
import { useNavigate } from "react-router-dom";

const PLACEHOLDER_MESSAGE =
  "✨ Want to refine this video?\n\nEdit functionality is coming soon! For now, you can:\n\n→ Create a refined version with your changes";

/**
 * ChatInterface renders the placeholder chat experience for refining videos.
 *
 * @param {{jobData: Object}} props - Component props
 * @returns {JSX.Element} Chat interface container
 */
function ChatInterface({ jobData }) {
  const navigate = useNavigate();
  const storageKey = useMemo(
    () => (jobData?.job_id ? `chat-history-${jobData.job_id}` : null),
    [jobData?.job_id]
  );

  const initialSystemMessage = useMemo(
    () => ({
      id: `system-welcome-${jobData?.job_id || "pending"}`,
      type: "system",
      text: PLACEHOLDER_MESSAGE,
    }),
    [jobData?.job_id]
  );

  const messages = useMemo(() => {
    if (!storageKey) {
      return [initialSystemMessage];
    }

    try {
      const storedMessages = window.localStorage.getItem(storageKey);
      if (storedMessages) {
        const parsedMessages = JSON.parse(storedMessages);
        const hasSystemMessage = parsedMessages.some((message) =>
          message.id?.startsWith("system-welcome")
        );
        return hasSystemMessage
          ? parsedMessages
          : [initialSystemMessage, ...parsedMessages];
      }
      return [initialSystemMessage];
    } catch {
      return [initialSystemMessage];
    }
  }, [initialSystemMessage, storageKey]);
  const [inputValue, setInputValue] = useState("");

  useEffect(() => {
    if (!storageKey || !messages.length) return;

    try {
      window.localStorage.setItem(storageKey, JSON.stringify(messages));
    } catch {
      // Silently ignore quota/storage issues
    }
  }, [messages, storageKey]);

  /**
   * Navigates to the Create page with the existing prompt pre-filled.
   */
  const handleCreateNewVideo = () => {
    navigate("/create", {
      state: {
        prefillPrompt: jobData?.prompt || "",
        sourceVideoId: jobData?.job_id || null,
        isRefinement: true, // Indicate this is a refinement of current video
      },
    });
  };

  /**
   * Prevents form submission until chat functionality ships.
   *
   * @param {Event} event - Submit event
   */
  const handleSubmit = (event) => {
    event.preventDefault();
  };

  return (
    <div className="chat-interface">
      <header className="chat-header">
        <div>
          <p className="chat-eyebrow">AI Assistant</p>
          <h2>Iterate with Chat</h2>
        </div>
        <span className="chat-badge" aria-live="polite">
          Coming Soon
        </span>
      </header>

      <div className="chat-messages" aria-live="polite">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`chat-message chat-message--${message.type}`}
          >
            <div className="chat-message-content">
              <p>{message.text}</p>
              {message.id.startsWith("system-welcome") && (
                <button
                  type="button"
                  className="btn-create-new"
                  onClick={handleCreateNewVideo}
                >
                  Refine This Video
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      <form className="chat-input-container" onSubmit={handleSubmit}>
        <textarea
          className="chat-input"
          placeholder="Describe how you'd like to refine this video..."
          value={inputValue}
          onChange={(event) => setInputValue(event.target.value)}
          disabled
          rows={3}
        />
        <button
          type="submit"
          className="chat-submit-btn"
          disabled
          title="Refine functionality coming in Phase 2"
        >
          Refine (Coming Soon)
        </button>
      </form>
    </div>
  );
}

ChatInterface.propTypes = {
  jobData: PropTypes.shape({
    job_id: PropTypes.string.isRequired,
    prompt: PropTypes.string,
  }).isRequired,
};

export default ChatInterface;
