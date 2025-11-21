package handlers

import "testing"

func TestS3KeyGenerationHelpers(t *testing.T) {
	userID := "user123"
	jobID := "job456"

	testCases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "scene clip key",
			got:  buildSceneClipKey(userID, jobID, 3),
			want: "users/user123/jobs/job456/clips/scene-003.mp4",
		},
		{
			name: "scene thumbnail key",
			got:  buildSceneThumbnailKey(userID, jobID, 5),
			want: "users/user123/jobs/job456/thumbnails/scene-005.jpg",
		},
		{
			name: "job thumbnail key",
			got:  buildJobThumbnailKey(userID, jobID),
			want: "users/user123/jobs/job456/thumbnails/job-thumbnail.jpg",
		},
		{
			name: "background music key",
			got:  buildAudioKey(userID, jobID),
			want: "users/user123/jobs/job456/audio/background-music.mp3",
		},
		{
			name: "narrator audio key",
			got:  buildNarratorAudioKey(userID, jobID),
			want: "users/user123/jobs/job456/audio/narrator-voiceover.mp3",
		},
		{
			name: "final video key",
			got:  buildFinalVideoKey(userID, jobID),
			want: "users/user123/jobs/job456/final/video.mp4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, tc.got)
			}
		})
	}
}
