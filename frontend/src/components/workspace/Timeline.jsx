import PropTypes from "prop-types";
import { Video, Music, Type } from "lucide-react";

/**
 * Timeline component for video editing with video, music, and text tracks
 * @param {Object} props - Component props
 */
function Timeline({ 
  videoDuration = 30, 
  scenes = [], 
  audioSpec = null, 
  audioUrl = null,
  onSeek 
}) {
  // Map scenes to video track segments
  const videoTrack = {
    segments: scenes.map((scene, index) => ({
      id: scene.scene_number || index + 1,
      start: scene.start_time || 0,
      end: (scene.start_time || 0) + (scene.duration || 0),
      label: scene.location || `Scene ${scene.scene_number || index + 1}`,
      action: scene.action,
    })),
  };

  // Map audio to music track (if audio is enabled and URL exists)
  const musicTrack = {
    segments: audioSpec?.enable_audio && audioUrl ? [
      {
        id: 1,
        start: 0,
        end: videoDuration,
        label: audioSpec.music_mood ? `${audioSpec.music_style || 'Music'} - ${audioSpec.music_mood}` : 'Background Music',
      },
    ] : [],
  };

  // Map voiceover/sync points to text track
  const textTrack = {
    segments: (() => {
      const segments = [];
      
      // Add voiceover text if available
      if (audioSpec?.voiceover_text) {
        // Split voiceover into segments based on sync points or evenly distribute
        if (audioSpec.sync_points && audioSpec.sync_points.length > 0) {
          audioSpec.sync_points.forEach((syncPoint, index) => {
            const nextPoint = audioSpec.sync_points[index + 1];
            segments.push({
              id: `voiceover-${index + 1}`,
              start: syncPoint.timestamp || 0,
              end: nextPoint ? nextPoint.timestamp : videoDuration,
              text: syncPoint.description || audioSpec.voiceover_text,
            });
          });
        } else {
          // Single voiceover segment spanning the duration
          segments.push({
            id: 'voiceover-1',
            start: 0,
            end: videoDuration,
            text: audioSpec.voiceover_text,
          });
        }
      }
      
      return segments;
    })(),
  };

  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  // Calculate minimum width for timeline content (e.g., 50px per second)
  const timelineContentWidth = videoDuration * 50;

  const renderSegment = (segment, trackWidth, totalDuration) => {
    const width = ((segment.end - segment.start) / totalDuration) * timelineContentWidth;
    const left = (segment.start / totalDuration) * timelineContentWidth;
    return (
      <div
        key={segment.id}
        className="timeline-segment"
        style={{
          width: `${width}px`,
          left: `${left}px`,
        }}
      >
        {segment.text && (
          <span className="timeline-segment-text">{segment.text}</span>
        )}
        {segment.label && (
          <span className="timeline-segment-label">{segment.label}</span>
        )}
      </div>
    );
  };

  // Generate multi-level time markers
  // Major markers every 5 seconds, minor every 1 second, sub-minor every 0.5 seconds
  const majorMarkers = Array.from(
    { length: Math.ceil(videoDuration / 5) + 1 },
    (_, i) => i * 5
  ).filter((time) => time <= videoDuration);

  const minorMarkers = Array.from(
    { length: Math.ceil(videoDuration) + 1 },
    (_, i) => i
  ).filter((time) => time <= videoDuration && !majorMarkers.includes(time));

  const subMinorMarkers = Array.from(
    { length: Math.ceil(videoDuration * 2) + 1 },
    (_, i) => i * 0.5
  ).filter((time) => time <= videoDuration && !majorMarkers.includes(time) && !minorMarkers.includes(time));

  return (
    <div className="timeline-container">
      <div className="timeline-scroll-wrapper">
        {/* Time Ruler */}
        <div className="timeline-ruler">
          <div className="timeline-ruler-spacer"></div>
          <div 
            className="timeline-ruler-content"
            style={{ minWidth: `${timelineContentWidth}px` }}
          >
            {/* Sub-minor markers (thin lines, no labels) */}
            {subMinorMarkers.map((time) => (
              <div
                key={`sub-${time}`}
                className="timeline-ruler-marker timeline-ruler-marker-subminor"
                style={{ left: `${(time / videoDuration) * timelineContentWidth}px` }}
              />
            ))}
            {/* Minor markers (medium lines, no labels) */}
            {minorMarkers.map((time) => (
              <div
                key={`minor-${time}`}
                className="timeline-ruler-marker timeline-ruler-marker-minor"
                style={{ left: `${(time / videoDuration) * timelineContentWidth}px` }}
              />
            ))}
            {/* Major markers (thick lines with labels) */}
            {majorMarkers.map((time) => (
              <div
                key={`major-${time}`}
                className="timeline-ruler-marker timeline-ruler-marker-major"
                style={{ left: `${(time / videoDuration) * timelineContentWidth}px` }}
              >
                <span className="timeline-ruler-label">{time}s</span>
              </div>
            ))}
          </div>
        </div>

        <div className="timeline-tracks">
        {/* Video Track */}
        <div className="timeline-track">
          <div className="timeline-track-header">
            <Video size={18} />
            <span className="timeline-track-label">Video</span>
          </div>
          <div 
            className="timeline-track-content timeline-track-content-video"
            style={{ minWidth: `${timelineContentWidth}px` }}
          >
            {videoTrack.segments.map((segment) =>
              renderSegment(segment, 100, videoDuration)
            )}
          </div>
        </div>

        {/* Music Track */}
        <div className="timeline-track">
          <div className="timeline-track-header">
            <Music size={18} />
            <span className="timeline-track-label">Music</span>
          </div>
          <div 
            className="timeline-track-content timeline-track-content-music"
            style={{ minWidth: `${timelineContentWidth}px` }}
          >
            {musicTrack.segments.map((segment) =>
              renderSegment(segment, 100, videoDuration)
            )}
          </div>
        </div>

        {/* Text Track */}
        <div className="timeline-track">
          <div className="timeline-track-header">
            <Type size={18} />
            <span className="timeline-track-label">Text</span>
          </div>
          <div 
            className="timeline-track-content timeline-track-content-text"
            style={{ minWidth: `${timelineContentWidth}px` }}
          >
            {textTrack.segments.map((segment) =>
              renderSegment(segment, 100, videoDuration)
            )}
          </div>
        </div>
      </div>
      </div>
    </div>
  );
}

Timeline.propTypes = {
  videoDuration: PropTypes.number,
  scenes: PropTypes.arrayOf(PropTypes.shape({
    scene_number: PropTypes.number,
    start_time: PropTypes.number,
    duration: PropTypes.number,
    location: PropTypes.string,
    action: PropTypes.string,
  })),
  audioSpec: PropTypes.shape({
    enable_audio: PropTypes.bool,
    music_mood: PropTypes.string,
    music_style: PropTypes.string,
    voiceover_text: PropTypes.string,
    sync_points: PropTypes.arrayOf(PropTypes.shape({
      timestamp: PropTypes.number,
      description: PropTypes.string,
    })),
  }),
  audioUrl: PropTypes.string,
  onSeek: PropTypes.func,
};

Timeline.defaultProps = {
  videoDuration: 30,
  scenes: [],
  audioSpec: null,
  audioUrl: null,
  onSeek: undefined,
};

export default Timeline;

