import { useEffect, useMemo, useRef, useState } from "react";
import PropTypes from "prop-types";

/**
 * ChatInterface renders a functional chat experience for refining videos.
 *
 * @param {{jobData: Object, onRefine: Function}} props - Component props
 * @returns {JSX.Element} Chat interface container
 */
function ChatInterface({ jobData, onRefine }) {
  const storageKey = useMemo(
    () => (jobData?.job_id ? `chat-history-${jobData.job_id}` : null),
    [jobData?.job_id]
  );

  const messagesEndRef = useRef(null);

  // Load messages from localStorage
  const [messages, setMessages] = useState(() => {
    if (!storageKey) return [];
    
    try {
      const storedMessages = window.localStorage.getItem(storageKey);
      if (storedMessages) {
        return JSON.parse(storedMessages);
      }
    } catch (error) {
      console.error("[CHAT] Failed to load chat history:", error);
    }
    return [];
  });

  const [inputValue, setInputValue] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Save messages to localStorage whenever they change
  useEffect(() => {
    if (!storageKey || !messages.length) return;

    try {
      window.localStorage.setItem(storageKey, JSON.stringify(messages));
    } catch (error) {
      console.error("[CHAT] Failed to save chat history:", error);
    }
  }, [messages, storageKey]);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  /**
   * Handles form submission for refinement requests.
   *
   * @param {Event} event - Submit event
   */
  const handleSubmit = async (event) => {
    event.preventDefault();
    
    const trimmedInput = inputValue.trim();
    if (!trimmedInput || isSubmitting) {
      return;
    }

    if (!onRefine) {
      console.warn("[CHAT] No onRefine handler provided");
      return;
    }

    setIsSubmitting(true);

    // Add user message to chat
    const userMessage = {
      id: `user-${Date.now()}-${Math.random()}`,
      type: "user",
      text: trimmedInput,
      timestamp: Date.now(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInputValue("");

    try {
      console.log("[CHAT] 🎬 Starting refinement with prompt:", trimmedInput);
      
      // Call the refinement handler from parent
      await onRefine(trimmedInput);

      // Add system confirmation message
      const systemMessage = {
        id: `system-${Date.now()}-${Math.random()}`,
        type: "system",
        text: "Refinement started! Your video is being generated...",
        timestamp: Date.now(),
      };

      setMessages((prev) => [...prev, systemMessage]);
    } catch (error) {
      console.error("[CHAT] ❌ Refinement failed:", error);
      
      // Add error message
      const errorMessage = {
        id: `error-${Date.now()}-${Math.random()}`,
        type: "system",
        text: `Failed to start refinement: ${error.message || "Unknown error"}`,
        timestamp: Date.now(),
        isError: true,
      };

      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="chat-interface">
      <header className="chat-header">
        <div>
          <p className="chat-eyebrow">AI Assistant</p>
          <h2>Iterate with Chat</h2>
        </div>
      </header>

      <div className="chat-messages" aria-live="polite">
        {messages.length === 0 ? (
          <div className="chat-message chat-message--system">
            <div className="chat-message-content">
              <p>Start a conversation to refine your video. Describe the changes you'd like to make.</p>
            </div>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`chat-message chat-message--${message.type} ${
                message.isError ? "chat-message--error" : ""
              }`}
            >
              <div className="chat-message-content">
                <p>{message.text}</p>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      <form className="chat-input-container" onSubmit={handleSubmit}>
        <textarea
          className="chat-input"
          placeholder="Describe how you'd like to refine this video..."
          value={inputValue}
          onChange={(event) => setInputValue(event.target.value)}
          disabled={isSubmitting}
          rows={3}
        />
        <button
          type="submit"
          className="chat-submit-btn"
          disabled={!inputValue.trim() || isSubmitting}
          title={isSubmitting ? "Processing..." : "Refine video"}
        >
          {isSubmitting ? "Processing..." : "Refine"}
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
  onRefine: PropTypes.func.isRequired,
};

export default ChatInterface;
