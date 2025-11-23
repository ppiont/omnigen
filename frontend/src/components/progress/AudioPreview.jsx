import PropTypes from "prop-types";

function AudioTrackCard({ title, description, url, status }) {
  return (
    <div className="audio-track-card">
      <div className="audio-track-info">
        <p className="audio-track-title">{title}</p>
        <p className="audio-track-description">
          {url ? description : status === "generating" ? "Generating..." : "Not available yet"}
        </p>
      </div>
      {url ? (
        <audio controls preload="metadata">
          <source src={url} type="audio/mpeg" />
          <source src={url} type="audio/wav" />
          Your browser does not support the audio element.
        </audio>
      ) : (
        <div className="audio-track-placeholder" aria-label={`${title} not ready`}>
          <span>{status === "generating" ? "Processing audio…" : "Pending"}</span>
        </div>
      )}
    </div>
  );
}

AudioTrackCard.propTypes = {
  title: PropTypes.string.isRequired,
  description: PropTypes.string,
  url: PropTypes.string,
  status: PropTypes.oneOf(["pending", "generating", "ready"]).isRequired,
};

function AudioPreview({ narratorUrl, backgroundUrl, isGeneratingNarrator, isGeneratingMusic }) {
  return (
    <section className="job-progress-section">
      <div className="section-heading">
        <h3>Audio Preview</h3>
        <p>Listen to narration and background music as soon as they are ready.</p>
      </div>
      <div className="audio-track-grid">
        <AudioTrackCard
          title="Narrator Voiceover"
          description="Generated FDA-compliant narration track"
          url={narratorUrl}
          status={narratorUrl ? "ready" : isGeneratingNarrator ? "generating" : "pending"}
        />
        <AudioTrackCard
          title="Background Music"
          description="Cinematic backing track"
          url={backgroundUrl}
          status={backgroundUrl ? "ready" : isGeneratingMusic ? "generating" : "pending"}
        />
      </div>
    </section>
  );
}

AudioPreview.propTypes = {
  narratorUrl: PropTypes.string,
  backgroundUrl: PropTypes.string,
  isGeneratingNarrator: PropTypes.bool,
  isGeneratingMusic: PropTypes.bool,
};

export default AudioPreview;

