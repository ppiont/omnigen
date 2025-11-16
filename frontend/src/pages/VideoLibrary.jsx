import { useNavigate } from "react-router-dom";
import VideoCard from "../components/VideoCard.jsx";
import "../styles/dashboard.css";

const recentVideos = [
  {
    id: "1",
    title: "Product Showcase - Tech Headphones",
    status: "Completed",
    createdAt: "2h ago",
    duration: "30s",
    aspectRatios: ["16:9", "9:16", "1:1"],
    thumbnailUrl:
      "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=600&h=400&fit=crop",
  },
  {
    id: "2",
    title: "Social Ad - Fashion Brand",
    status: "Completed",
    createdAt: "4h ago",
    duration: "15s",
    aspectRatios: ["9:16"],
    thumbnailUrl:
      "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=600&h=400&fit=crop",
  },
  {
    id: "3",
    title: "Explainer Video - SaaS Platform",
    status: "Processing",
    createdAt: "1d ago",
    duration: "60s",
    aspectRatios: ["16:9"],
    thumbnailUrl:
      "https://images.unsplash.com/photo-1551650975-87deedd944c3?w=600&h=400&fit=crop",
  },
  {
    id: "4",
    title: "Instagram Story - Coffee Brand",
    status: "Completed",
    createdAt: "2d ago",
    duration: "15s",
    aspectRatios: ["9:16", "1:1"],
    thumbnailUrl:
      "https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=600&h=400&fit=crop",
  },
  {
    id: "5",
    title: "Product Demo - Smart Watch",
    status: "Completed",
    createdAt: "3d ago",
    duration: "30s",
    aspectRatios: ["1:1", "9:16"],
    thumbnailUrl:
      "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=600&h=400&fit=crop",
  },
  {
    id: "6",
    title: "Brand Story - Eco Friendly",
    status: "Completed",
    createdAt: "5d ago",
    duration: "90s",
    aspectRatios: ["16:9", "9:16", "1:1"],
    thumbnailUrl:
      "https://images.unsplash.com/photo-1542601906990-b4d3fb778b09?w=600&h=400&fit=crop",
  },
];

function VideoLibrary() {
  const navigate = useNavigate();

  const handleVideoClick = (videoId) => {
    navigate(`/workspace/${videoId}`);
  };

  return (
    <div className="video-library-page">
      <h1 className="page-title">Video Library</h1>
      <section className="videos-section">
        <h2 className="section-title">Your videos</h2>
        {recentVideos.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              Your generated videos will appear here
            </p>
          </div>
        ) : (
          <div className="videos-grid">
            {recentVideos.map((video) => (
              <div
                key={video.id}
                onClick={() => handleVideoClick(video.id)}
                style={{ cursor: "pointer" }}
              >
                <VideoCard
                  video={video}
                  onDownload={() => console.log('Download', video.id)}
                  onDelete={() => console.log('Delete', video.id)}
                />
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export default VideoLibrary;
