import PropTypes from "prop-types";
import { Video, Music, Type, Mic } from "lucide-react";

/**
 * Timeline component for video editing with video, music, and text tracks
 * @param {Object} props - Component props
 */
function Timeline({
  videoDuration = 30,
  scenes = [],
  audioSpec = null,
  backgroundMusicUrl = null,
  narratorAudioUrl = null,
  sideEffectsText = null,
  sideEffectsStartTime = null,
  audioUrl = null,
  onSeek,
}) {
  // Debug logging
  console.log("[TIMELINE] Received props:", {
    videoDuration,
    scenesCount: scenes.length,
    scenes: scenes,
    audioSpec: audioSpec,
    backgroundMusicUrl,
    narratorAudioUrl,
    sideEffectsText,
    sideEffectsStartTime,
    deprecatedAudioUrl: audioUrl,
  });

  // Map scenes to video track segments
  const videoTrack = {
    segments: scenes.map((scene, index) => {
      const segment = {
        id: scene.scene_number || scene.sceneNumber || index + 1,
        start: scene.start_time || scene.startTime || 0,
        end: (scene.start_time || scene.startTime || 0) + (scene.duration || 0),
        label:
          scene.location ||
          `Scene ${scene.scene_number || scene.sceneNumber || index + 1}`,
        action: scene.action,
      };
      console.log(`[TIMELINE] Mapped scene ${index + 1}:`, segment);
      return segment;
    }),
  };

  const resolvedMusicUrl = backgroundMusicUrl || audioUrl;

  // Map audio to music track (if audio is enabled and URL exists)
  const musicTrack = {
    segments:
      audioSpec?.enable_audio && resolvedMusicUrl
        ? [
            {
              id: 1,
              start: 0,
              end: videoDuration,
              label: audioSpec.music_mood
                ? `${audioSpec.music_style || "Music"} - ${
                    audioSpec.music_mood
                  }`
                : "Background Music",
            },
          ]
        : [],
  };

  // Map narrator voiceover audio to audio track
  const audioTrack = {
    segments: narratorAudioUrl
      ? [
          {
            id: 1,
            start: 0,
            end: videoDuration,
            label: "Narrator Voiceover",
          },
        ]
      : [],
  };

  const resolvedSideEffectsStart =
    typeof sideEffectsStartTime === "number"
      ? sideEffectsStartTime
      : sideEffectsText
      ? videoDuration * 0.8
      : null;

  const truncatedSideEffectsText =
    sideEffectsText && sideEffectsText.length > 80
      ? `${sideEffectsText.slice(0, 77)}...`
      : sideEffectsText || null;

  // Map side effects overlay to text track
  const textTrack = {
    segments:
      truncatedSideEffectsText && typeof resolvedSideEffectsStart === "number"
        ? [
            {
              id: "side-effects",
              start: Math.max(
                0,
                Math.min(resolvedSideEffectsStart, videoDuration)
              ),
              end: videoDuration,
              text: truncatedSideEffectsText,
            },
          ]
        : [],
  };

  console.log("[TIMELINE] Video track segments:", videoTrack.segments);
  console.log("[TIMELINE] Music track segments:", musicTrack.segments);
  console.log("[TIMELINE] Audio track segments:", audioTrack.segments);
  console.log("[TIMELINE] Text track segments:", textTrack.segments);

  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  // For videos <= 30 seconds: fill the timeline (100% width)
  // For videos > 30 seconds: use calculated width (50px per second) and allow horizontal scroll
  const isShortVideo = videoDuration <= 30;
  const timelineContentWidth = isShortVideo ? null : videoDuration * 50; // null means use 100%

  const renderSegment = (segment, trackWidth, totalDuration) => {
    if (isShortVideo) {
      // Use percentage-based positioning for short videos
      const widthPercent =
        ((segment.end - segment.start) / totalDuration) * 100;
      const leftPercent = (segment.start / totalDuration) * 100;
      return (
        <div
          key={segment.id}
          className="timeline-segment"
          style={{
            width: `${widthPercent}%`,
            left: `${leftPercent}%`,
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
    } else {
      // Use pixel-based positioning for longer videos
      const width =
        ((segment.end - segment.start) / totalDuration) * timelineContentWidth;
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
    }
  };

  // Generate multi-level time markers with dynamic intervals
  // Calculate appropriate interval based on video duration
  const getMajorInterval = (duration) => {
    if (duration <= 10) return 2; // Every 2 seconds for very short videos
    if (duration <= 30) return 5; // Every 5 seconds for short videos
    if (duration <= 120) return 10; // Every 10 seconds for medium videos
    if (duration <= 300) return 30; // Every 30 seconds for long videos
    return 60; // Every 60 seconds for very long videos
  };

  const getMinorInterval = (duration) => {
    if (duration <= 10) return 1; // Every 1 second for very short videos
    if (duration <= 30) return 1; // Every 1 second for short videos
    if (duration <= 120) return 5; // Every 5 seconds for medium videos
    return 10; // Every 10 seconds for longer videos
  };

  const majorInterval = getMajorInterval(videoDuration);
  const minorInterval = getMinorInterval(videoDuration);

  // Generate major markers - always include 0 and videoDuration
  const majorMarkers = [];
  for (let time = 0; time <= videoDuration; time += majorInterval) {
    if (time <= videoDuration) {
      majorMarkers.push(time);
    }
  }
  // Ensure the final marker is exactly at videoDuration (remove duplicates and sort)
  const uniqueMajorMarkers = [...new Set(majorMarkers)];
  if (!uniqueMajorMarkers.includes(videoDuration)) {
    uniqueMajorMarkers.push(videoDuration);
  }
  uniqueMajorMarkers.sort((a, b) => a - b);
  const finalMajorMarkers = uniqueMajorMarkers;

  // Generate minor markers (exclude major markers)
  const minorMarkers = [];
  for (let time = 0; time <= videoDuration; time += minorInterval) {
    if (time <= videoDuration && !finalMajorMarkers.includes(time)) {
      minorMarkers.push(time);
    }
  }

  // Generate sub-minor markers (every 0.5 seconds, exclude major and minor)
  const subMinorMarkers = [];
  for (let time = 0; time <= videoDuration; time += 0.5) {
    if (
      time <= videoDuration &&
      !finalMajorMarkers.includes(time) &&
      !minorMarkers.includes(time)
    ) {
      subMinorMarkers.push(time);
    }
  }

  return (
    <div className="timeline-container">
      <div className="timeline-scroll-wrapper">
        {/* Time Ruler */}
        <div className="timeline-ruler">
          <div className="timeline-ruler-spacer"></div>
          <div
            className="timeline-ruler-content"
            style={
              isShortVideo
                ? { width: "100%" }
                : {
                    width: `${timelineContentWidth}px`,
                    minWidth: `${timelineContentWidth}px`,
                  }
            }
          >
            {/* Sub-minor markers (thin lines, no labels) */}
            {subMinorMarkers.map((time) => (
              <div
                key={`sub-${time}`}
                className="timeline-ruler-marker timeline-ruler-marker-subminor"
                style={
                  isShortVideo
                    ? { left: `${(time / videoDuration) * 100}%` }
                    : {
                        left: `${
                          (time / videoDuration) * timelineContentWidth
                        }px`,
                      }
                }
              />
            ))}
            {/* Minor markers (medium lines, no labels) */}
            {minorMarkers.map((time) => (
              <div
                key={`minor-${time}`}
                className="timeline-ruler-marker timeline-ruler-marker-minor"
                style={
                  isShortVideo
                    ? { left: `${(time / videoDuration) * 100}%` }
                    : {
                        left: `${
                          (time / videoDuration) * timelineContentWidth
                        }px`,
                      }
                }
              />
            ))}
            {/* Major markers (thick lines with labels) */}
            {finalMajorMarkers.map((time) => (
              <div
                key={`major-${time}`}
                className="timeline-ruler-marker timeline-ruler-marker-major"
                style={
                  isShortVideo
                    ? { left: `${(time / videoDuration) * 100}%` }
                    : {
                        left: `${
                          (time / videoDuration) * timelineContentWidth
                        }px`,
                      }
                }
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
              style={
                isShortVideo
                  ? { width: "100%" }
                  : {
                      width: `${timelineContentWidth}px`,
                      minWidth: `${timelineContentWidth}px`,
                    }
              }
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
              style={
                isShortVideo
                  ? { width: "100%" }
                  : {
                      width: `${timelineContentWidth}px`,
                      minWidth: `${timelineContentWidth}px`,
                    }
              }
            >
              {musicTrack.segments.map((segment) =>
                renderSegment(segment, 100, videoDuration)
              )}
            </div>
          </div>

          {/* Audio Track */}
          <div className="timeline-track">
            <div className="timeline-track-header">
              <Mic size={18} />
              <span className="timeline-track-label">Audio</span>
            </div>
            <div
              className="timeline-track-content timeline-track-content-audio"
              style={
                isShortVideo
                  ? { width: "100%" }
                  : {
                      width: `${timelineContentWidth}px`,
                      minWidth: `${timelineContentWidth}px`,
                    }
              }
            >
              {audioTrack.segments.map((segment) =>
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
              style={
                isShortVideo
                  ? { width: "100%" }
                  : {
                      width: `${timelineContentWidth}px`,
                      minWidth: `${timelineContentWidth}px`,
                    }
              }
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
  scenes: PropTypes.arrayOf(
    PropTypes.shape({
      scene_number: PropTypes.number,
      start_time: PropTypes.number,
      duration: PropTypes.number,
      location: PropTypes.string,
      action: PropTypes.string,
    })
  ),
  audioSpec: PropTypes.shape({
    enable_audio: PropTypes.bool,
    music_mood: PropTypes.string,
    music_style: PropTypes.string,
  }),
  backgroundMusicUrl: PropTypes.string,
  narratorAudioUrl: PropTypes.string,
  sideEffectsText: PropTypes.string,
  sideEffectsStartTime: PropTypes.number,
  audioUrl: PropTypes.string,
  onSeek: PropTypes.func,
};

Timeline.defaultProps = {
  videoDuration: 30,
  scenes: [],
  audioSpec: null,
  backgroundMusicUrl: null,
  narratorAudioUrl: null,
  sideEffectsText: null,
  sideEffectsStartTime: null,
  audioUrl: null,
  onSeek: undefined,
};

export default Timeline;
