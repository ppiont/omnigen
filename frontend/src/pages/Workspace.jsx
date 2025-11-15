import { useParams, useNavigate, Link } from "react-router-dom";
import "../styles/workspace.css";

function Workspace() {
  const { videoId } = useParams();
  const navigate = useNavigate();

  // UUID validation (basic check)
  const isValidUUID = (id) => {
    const uuidRegex =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return uuidRegex.test(id);
  };

  // Validate video ID
  if (!videoId || !isValidUUID(videoId)) {
    return (
      <div className="workspace-error">
        <h1>Invalid Video ID</h1>
        <p>The video ID is invalid or missing.</p>
        <Link to="/library" className="btn btn-primary">
          Go to Library
        </Link>
      </div>
    );
  }

  return (
    <div className="workspace-page">
      {/* Breadcrumbs */}
      <nav className="workspace-breadcrumbs" aria-label="Breadcrumb">
        <Link to="/library" className="breadcrumb-link">
          Library
        </Link>
        <span className="breadcrumb-separator"> / </span>
        <span className="breadcrumb-current">Video Editor</span>
      </nav>

      {/* Placeholder Content */}
      <main className="workspace-main">
        <h1 className="workspace-title">Video Workspace</h1>
        <p className="workspace-video-id">
          Video ID: <code>{videoId}</code>
        </p>
        <p className="placeholder-text">
          This is a placeholder for the video editor workspace. Team members
          will implement the full editor functionality here.
        </p>

        {/* Back to Library Button */}
        <button
          onClick={() => navigate("/library")}
          className="btn btn-secondary"
        >
          ← Back to Library
        </button>
      </main>
    </div>
  );
}

export default Workspace;
