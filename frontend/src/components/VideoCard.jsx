function VideoCard({ thumbnail, title, format, duration, cost, timestamp }) {
  return (
    <article className="video-card">
      <div className="video-thumbnail">
        <img src={thumbnail} alt={title} />
        <span className="format-badge">{format}</span>
      </div>
      <div className="video-info">
        <h3 className="video-title">{title}</h3>
        <div className="video-meta">
          <span className="video-duration">{duration}</span>
          <span className="video-cost">{cost}</span>
          <span className="video-timestamp">{timestamp}</span>
        </div>
      </div>
    </article>
  );
}

export default VideoCard;

