import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import JobProgressPreview from "../components/progress/JobProgressPreview.jsx";
import { useJobProgress } from "../hooks/useJobProgress.js";
import { jobs } from "../utils/api.js";
import { showToast } from "../utils/toast.js";
import "../styles/dashboard.css";

function JobDetails() {
  const { jobId } = useParams();
  const navigate = useNavigate();
  const [jobData, setJobData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const loadJob = useCallback(async () => {
    if (!jobId) {
      setError("Missing job identifier");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const data = await jobs.get(jobId);
      setJobData(data);
    } catch (err) {
      console.error("Failed to load job details", err);
      setError(err?.message || "Unable to load job details.");
    } finally {
      setLoading(false);
    }
  }, [jobId]);

  useEffect(() => {
    loadJob();
  }, [loadJob]);

  const jobProgress = useJobProgress(jobId, {
    onComplete: () => {
      showToast("Video generation completed", "success");
      loadJob();
    },
    onFailed: () => {
      showToast("Video generation failed", "error");
      loadJob();
    },
  });

  const mergedJob = useMemo(() => {
    if (!jobData) {
      return jobData;
    }

    // Merge latest asset information from progress stream with static job payload
    if (jobProgress.progress?.job_id === jobData.job_id) {
      return {
        ...jobData,
        ...jobProgress.progress,
        scenes: jobData.scenes || jobProgress.progress.scenes || jobData.metadata?.script?.scenes,
        scene_video_urls:
          jobProgress.progress.scene_video_urls ||
          jobData.scene_video_urls,
      };
    }

    return jobData;
  }, [jobData, jobProgress.progress]);

  const handleViewWorkspace = () => {
    if (jobId) navigate(`/workspace/${jobId}`);
  };

  const handleBack = () => {
    navigate(-1);
  };

  const renderContent = () => {
    if (loading) {
      return (
        <div className="empty-state">
          <p className="empty-state-text">Loading job details…</p>
        </div>
      );
    }

    if (error) {
      return (
        <div className="empty-state">
          <p className="empty-state-text">{error}</p>
          <button className="btn-primary" onClick={loadJob}>
            Try again
          </button>
        </div>
      );
    }

    if (!mergedJob) {
      return (
        <div className="empty-state">
          <p className="empty-state-text">Job was not found.</p>
        </div>
      );
    }

    return (
      <JobProgressPreview
        job={mergedJob}
        progress={jobProgress.progress}
        percentage={jobProgress.percentage}
        stageTimeline={jobProgress.stageTimeline}
        estimatedTimeRemaining={jobProgress.estimatedTimeRemaining}
        onRefresh={loadJob}
      />
    );
  };

  return (
    <div className="job-details-page">
      <div className="page-actions">
      <button type="button" className="btn-ghost" onClick={handleBack}>
        ← Back
      </button>
      {mergedJob?.status === "completed" ? (
        <button type="button" className="btn-primary" onClick={handleViewWorkspace}>
          Open in Workspace
        </button>
      ) : null}
      </div>
      {renderContent()}
    </div>
  );
}

export default JobDetails;
