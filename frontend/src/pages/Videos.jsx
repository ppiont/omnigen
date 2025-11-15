import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { Search, Filter, Play, Calendar, X } from "lucide-react";
import AppLayout from "../components/AppLayout.jsx";
import VideoCard from "../components/VideoCard.jsx";
import "../styles/videos.css";

const mockVideos = [
  {
    id: "vid-101",
    title: "Product Demo – Summer Sale",
    status: "Completed",
    createdAt: "2025-11-12T09:45:00Z",
    duration: "00:30",
    aspectRatios: ["16:9", "9:16"],
  },
  {
    id: "vid-102",
    title: "Luxury Watch – Holiday Campaign",
    status: "Processing",
    createdAt: "2025-11-13T15:10:00Z",
    duration: "00:45",
    aspectRatios: ["16:9"],
  },
  {
    id: "vid-103",
    title: "Energy Drink TikTok Burst",
    status: "Completed",
    createdAt: "2025-10-30T11:05:00Z",
    duration: "00:15",
    aspectRatios: ["9:16"],
  },
  {
    id: "vid-104",
    title: "Minimal Skincare Explainer",
    status: "Failed",
    createdAt: "2025-11-08T21:22:00Z",
    duration: "01:00",
    aspectRatios: ["16:9", "1:1"],
  },
  {
    id: "vid-105",
    title: "Tech Headphones Launch",
    status: "Completed",
    createdAt: "2025-11-02T08:00:00Z",
    duration: "00:30",
    aspectRatios: ["16:9"],
  },
  {
    id: "vid-106",
    title: "Eco-Friendly Brand Story",
    status: "Processing",
    createdAt: "2025-11-14T13:15:00Z",
    duration: "00:60",
    aspectRatios: ["16:9", "9:16", "1:1"],
  },
  {
    id: "vid-107",
    title: "Sneaker Drop – Social Ad",
    status: "Completed",
    createdAt: "2025-11-10T19:30:00Z",
    duration: "00:20",
    aspectRatios: ["9:16"],
  },
  {
    id: "vid-108",
    title: "Premium Coffee Storyboard",
    status: "Failed",
    createdAt: "2025-10-28T17:40:00Z",
    duration: "00:45",
    aspectRatios: ["16:9"],
  },
];

const statusOptions = [
  { label: "All", value: "all" },
  { label: "Completed", value: "completed" },
  { label: "Processing", value: "processing" },
  { label: "Failed", value: "failed" },
];
const sortOptions = [
  { value: "newest", label: "Newest" },
  { value: "oldest", label: "Oldest" },
  { value: "name-asc", label: "Name A-Z" },
  { value: "name-desc", label: "Name Z-A" },
];

function Videos() {
  const [videos, setVideos] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchInput, setSearchInput] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [sortBy, setSortBy] = useState("newest");
  const [dateRange, setDateRange] = useState({ from: "", to: "" });
  const [confirmModal, setConfirmModal] = useState({ isOpen: false, video: null });
  const [toasts, setToasts] = useState([]);
  const searchDebounceRef = useRef();

  useEffect(() => {
    const timer = setTimeout(() => {
      setVideos(mockVideos);
      setIsLoading(false);
    }, 800);

    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (searchDebounceRef.current) {
      clearTimeout(searchDebounceRef.current);
    }
    searchDebounceRef.current = setTimeout(() => {
      setSearchQuery(searchInput.trim());
    }, 300);
    return () => clearTimeout(searchDebounceRef.current);
  }, [searchInput]);

  const normalizeStatus = (status) => status?.toLowerCase() || "processing";

  const statusCounts = useMemo(() => {
    return videos.reduce(
      (acc, video) => {
        acc.total += 1;
        const key = normalizeStatus(video.status);
        acc[key] = (acc[key] || 0) + 1;
        return acc;
      },
      { total: 0, completed: 0, processing: 0, failed: 0 }
    );
  }, [videos]);

  const filteredVideos = useMemo(() => {
    let results = [...videos];

    if (searchQuery) {
      const lowered = searchQuery.toLowerCase();
      results = results.filter((video) => video.title.toLowerCase().includes(lowered));
    }

    if (statusFilter !== "all") {
      results = results.filter(
        (video) => normalizeStatus(video.status) === statusFilter
      );
    }

    if (dateRange.from) {
      results = results.filter(
        (video) => new Date(video.createdAt) >= new Date(dateRange.from)
      );
    }

    if (dateRange.to) {
      results = results.filter(
        (video) => new Date(video.createdAt) <= new Date(dateRange.to)
      );
    }

    results.sort((a, b) => {
      switch (sortBy) {
        case "oldest":
          return new Date(a.createdAt) - new Date(b.createdAt);
        case "name-asc":
          return a.title.localeCompare(b.title);
        case "name-desc":
          return b.title.localeCompare(a.title);
        case "newest":
        default:
          return new Date(b.createdAt) - new Date(a.createdAt);
      }
    });

    return results;
  }, [videos, searchQuery, statusFilter, sortBy, dateRange]);

  const clearFilters = () => {
    setSearchInput("");
    setSearchQuery("");
    setStatusFilter("all");
    setSortBy("newest");
    setDateRange({ from: "", to: "" });
  };

  const formatDate = (dateString) =>
    new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });

  const addToast = (message, variant = "default") => {
    const id = `${Date.now()}-${Math.random()}`;
    setToasts((prev) => [...prev, { id, message, variant }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((toast) => toast.id !== id));
    }, 3000);
  };

  const handleDownload = (video) => {
    addToast(`Downloading ${video.title}...`, "info");
    setTimeout(() => {
      addToast("Download complete!", "success");
    }, 1000);
  };

  const handleDelete = (video) => {
    setConfirmModal({ isOpen: true, video });
  };

  const closeModal = () => setConfirmModal({ isOpen: false, video: null });

  const confirmDelete = () => {
    if (!confirmModal.video) return;
    setVideos((prev) => prev.filter((video) => video.id !== confirmModal.video.id));
    addToast("Video deleted", "warning");
    closeModal();
  };

  const activeFilters =
    searchQuery ||
    statusFilter !== "all" ||
    sortBy !== "newest" ||
    dateRange.from ||
    dateRange.to;

  const resultCountLabel = isLoading
    ? "..."
    : `Showing ${filteredVideos.length} of ${statusCounts.total} videos`;

  return (
    <AppLayout>
      <section className="videos-page">
        <div className="videos-header">
          <div>
            <p className="eyebrow">Library</p>
            <h2 className="section-title">
              Video Library <span className="result-count">{resultCountLabel}</span>
            </h2>
            <p className="section-description">
              Browse, filter, and manage all of your generated videos in one
              place.
            </p>
          </div>
          <Link to="/create" className="primary-link-button">
            Create Video
          </Link>
        </div>

        <div className="videos-filters">
          <div className="filter-input">
            <Search size={16} />
            <input
              type="text"
              placeholder="Search videos..."
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
            />
            {searchInput && (
              <button
                type="button"
                className="clear-search"
                onClick={() => {
                  setSearchInput("");
                  setSearchQuery("");
                }}
                aria-label="Clear search"
              >
                <X size={14} />
              </button>
            )}
          </div>

          <div className="filter-select">
            <label htmlFor="status-filter">Status</label>
            <select
              id="status-filter"
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value)}
            >
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {`${option.label} ${
                    option.value === "all"
                      ? `(${statusCounts.total})`
                      : `(${statusCounts[option.value] || 0})`
                  }`}
                </option>
              ))}
            </select>
          </div>

          <div className="filter-daterange">
            <label htmlFor="date-from">Date range</label>
            <div className="date-inputs">
              <div className="date-input">
                <Calendar size={14} />
                <input
                  id="date-from"
                  type="date"
                  value={dateRange.from}
                  onChange={(event) =>
                    setDateRange((prev) => ({ ...prev, from: event.target.value }))
                  }
                />
              </div>
              <span className="date-separator">–</span>
              <div className="date-input">
                <Calendar size={14} />
                <input
                  type="date"
                  value={dateRange.to}
                  onChange={(event) =>
                    setDateRange((prev) => ({ ...prev, to: event.target.value }))
                  }
                />
              </div>
            </div>
          </div>

          <div className="filter-select">
            <label htmlFor="sort-by">Sort by</label>
            <select
              id="sort-by"
              value={sortBy}
              onChange={(event) => setSortBy(event.target.value)}
            >
              {sortOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          <button className="secondary-filter-button" type="button">
            <Filter size={16} />
            Advanced filters
          </button>

          {activeFilters && (
            <button
              type="button"
              className="clear-filters"
              onClick={clearFilters}
              aria-label="Clear all filters"
            >
              Clear filters
            </button>
          )}
        </div>

        {isLoading ? (
          <div className="videos-grid">
            {Array.from({ length: 6 }).map((_, index) => (
              <div className="video-card skeleton" key={`skeleton-${index}`}>
                <div className="video-thumbnail" />
                <div className="video-card-body">
                  <div className="skeleton-line w-60" />
                  <div className="skeleton-line w-40" />
                  <div className="skeleton-line w-80" />
                </div>
              </div>
            ))}
          </div>
        ) : filteredVideos.length === 0 ? (
          <div className="videos-empty">
            <div className="empty-illustration">
              <Play size={32} />
            </div>
            <h3>
              {searchQuery ? `No results for “${searchQuery}”` : "No videos found"}
            </h3>
            <p>Try adjusting your filters or create a new generation.</p>
            <Link to="/create" className="primary-link-button">
              Create your first video
            </Link>
          </div>
        ) : (
          <div className="videos-grid">
            {filteredVideos.map((video) => (
              <VideoCard
                key={video.id}
                video={{
                  ...video,
                  createdAt: formatDate(video.createdAt),
                }}
                onDownload={handleDownload}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
      </section>

      <ToastStack toasts={toasts} />
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        video={confirmModal.video}
        onCancel={closeModal}
        onConfirm={confirmDelete}
      />
    </AppLayout>
  );
}

function ToastStack({ toasts }) {
  return (
    <div className="toast-stack" aria-live="polite" aria-atomic="true">
      {toasts.map((toast) => (
        <div key={toast.id} className={`toast ${toast.variant}`}>
          {toast.message}
        </div>
      ))}
    </div>
  );
}

function ConfirmModal({ isOpen, video, onCancel, onConfirm }) {
  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        onCancel();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onCancel]);

  if (!isOpen || !video) return null;

  return (
    <div className="modal-backdrop" onClick={onCancel} role="dialog" aria-modal="true">
      <div className="modal-card" onClick={(event) => event.stopPropagation()}>
        <h3>Delete video</h3>
        <p>Delete “{video.title}”? This cannot be undone.</p>
        <div className="modal-actions">
          <button type="button" onClick={onCancel} className="modal-cancel">
            Cancel
          </button>
          <button type="button" onClick={onConfirm} className="modal-delete">
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

export default Videos;
