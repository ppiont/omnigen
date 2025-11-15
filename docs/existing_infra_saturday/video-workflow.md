# Video Generation Workflow

> Step Functions Express state machine orchestrating Lambda-based video generation pipeline

## Overview

The video generation pipeline uses **AWS Step Functions Express** to orchestrate a multi-stage workflow:
1. **Scene Generation** (parallel): Generate 5 video clips from AI models
2. **Video Composition** (sequential): Stitch clips with transitions and audio sync
3. **Post-Processing** (sequential): Update DynamoDB, increment usage quotas

**Workflow Type:** Express (synchronous, < 5 minutes)
**Max Execution Time:** 300 seconds (5 minutes)
**Cost:** ~$0.000025 per execution (50% cheaper than Standard)
**Concurrency:** Unlimited (limited by Lambda concurrency)

---

## Complete State Machine

This state diagram shows the entire Step Functions workflow with all states and transitions.

```mermaid
stateDiagram-v2
    [*] --> ValidateInput: Start Execution

    ValidateInput --> ParsePrompt: Input Valid
    ValidateInput --> HandleError: Input Invalid

    ParsePrompt --> PlanScenes: Prompt Parsed
    ParsePrompt --> HandleError: Parse Failed

    PlanScenes --> GenerateScenes: Scenes Planned
    PlanScenes --> HandleError: Planning Failed

    state GenerateScenes {
        [*] --> MapScenes: Parallel Execution

        state MapScenes {
            [*] --> Scene1
            [*] --> Scene2
            [*] --> Scene3
            [*] --> Scene4
            [*] --> Scene5

            Scene1 --> FetchAPIKey1
            Scene2 --> FetchAPIKey2
            Scene3 --> FetchAPIKey3
            Scene4 --> FetchAPIKey4
            Scene5 --> FetchAPIKey5

            FetchAPIKey1 --> CallReplicate1
            FetchAPIKey2 --> CallReplicate2
            FetchAPIKey3 --> CallReplicate3
            FetchAPIKey4 --> CallReplicate4
            FetchAPIKey5 --> CallReplicate5

            CallReplicate1 --> PollStatus1
            CallReplicate2 --> PollStatus2
            CallReplicate3 --> PollStatus3
            CallReplicate4 --> PollStatus4
            CallReplicate5 --> PollStatus5

            PollStatus1 --> UploadS31
            PollStatus2 --> UploadS32
            PollStatus3 --> UploadS33
            PollStatus4 --> UploadS34
            PollStatus5 --> UploadS35

            UploadS31 --> [*]
            UploadS32 --> [*]
            UploadS33 --> [*]
            UploadS34 --> [*]
            UploadS35 --> [*]
        }

        MapScenes --> [*]: All Scenes Complete
    }

    GenerateScenes --> ComposeVideo: All Scenes Generated
    GenerateScenes --> HandleError: Scene Generation Failed

    ComposeVideo --> DownloadScenes: Compose Started
    DownloadScenes --> RunFFmpeg: Scenes Downloaded
    RunFFmpeg --> UploadFinalVideo: FFmpeg Complete
    UploadFinalVideo --> UpdateJobStatus: Upload Complete

    ComposeVideo --> HandleError: Composition Failed
    DownloadScenes --> HandleError: Download Failed
    RunFFmpeg --> HandleError: FFmpeg Failed
    UploadFinalVideo --> HandleError: Upload Failed

    UpdateJobStatus --> IncrementUsage: Job Updated
    UpdateJobStatus --> HandleError: Update Failed

    IncrementUsage --> SendNotification: Usage Incremented
    SendNotification --> [*]: Workflow Complete

    IncrementUsage --> HandleError: Increment Failed
    SendNotification --> HandleError: Notification Failed

    HandleError --> LogError: Capture Error
    LogError --> UpdateJobFailed: Log Written
    UpdateJobFailed --> [*]: Workflow Failed

    note right of GenerateScenes
        Parallel Execution (Map State)
        5 concurrent Lambda invocations
        Max parallelism: 5
    end note

    note right of ComposeVideo
        Sequential Execution
        Single Lambda with FFmpeg
        10 GB memory, 900s timeout
    end note

    note right of HandleError
        All errors caught and logged
        Job marked as "failed" in DynamoDB
        User receives error notification
    end note
```

---

## State Machine Definition (JSON)

High-level structure of the Step Functions ASL (Amazon States Language) definition.

```mermaid
flowchart TB
    Start([StartAt: ValidateInput])

    subgraph ValidateInput[\"Task: ValidateInput\"]
        VI_Lambda[Lambda: ValidateInputFunction]
        VI_Retry[Retry: 2 attempts, exponential backoff]
        VI_Catch[Catch: States.ALL → HandleError]
    end

    subgraph ParsePrompt[\"Task: ParsePrompt\"]
        PP_Lambda[Lambda: ParsePromptFunction]
        PP_Retry[Retry: 2 attempts]
        PP_Catch[Catch: → HandleError]
    end

    subgraph PlanScenes[\"Task: PlanScenes\"]
        PS_Lambda[Lambda: PlanScenesFunction]
        PS_Retry[Retry: 2 attempts]
        PS_Catch[Catch: → HandleError]
    end

    subgraph GenerateScenes[\"Map: GenerateScenes\"]
        GS_ItemsPath[ItemsPath: $.scenes]
        GS_MaxConcurrency[MaxConcurrency: 5]
        GS_Iterator{Iterator State Machine}

        subgraph Iterator[\"Iterator\"]
            IT_GenerateScene[Task: GenerateSceneFunction]
            IT_Retry[Retry: 3 attempts, exponential backoff]
            IT_Catch[Catch: → HandleError]
        end
    end

    subgraph ComposeVideo[\"Task: ComposeVideo\"]
        CV_Lambda[Lambda: ComposeVideoFunction]
        CV_Timeout[Timeout: 180 seconds]
        CV_Retry[Retry: 1 attempt]
        CV_Catch[Catch: → HandleError]
    end

    subgraph UpdateJobStatus[\"Task: UpdateJobStatus\"]
        UJ_DynamoDB[DynamoDB UpdateItem]
        UJ_Retry[Retry: 3 attempts]
        UJ_Catch[Catch: → HandleError]
    end

    subgraph IncrementUsage[\"Task: IncrementUsage\"]
        IU_DynamoDB[DynamoDB UpdateItem<br/>Increment videosGenerated]
        IU_Retry[Retry: 3 attempts]
        IU_Catch[Catch: → HandleError]
    end

    subgraph HandleError[\"Task: HandleError\"]
        HE_Lambda[Lambda: ErrorHandlerFunction]
        HE_LogError[Log to CloudWatch]
        HE_UpdateDDB[Update Job Status: failed]
        HE_SNS[Send SNS Notification optional]
    end

    Start --> ValidateInput
    ValidateInput -->|Next| ParsePrompt
    ParsePrompt -->|Next| PlanScenes
    PlanScenes -->|Next| GenerateScenes
    GenerateScenes -->|Next| ComposeVideo
    ComposeVideo -->|Next| UpdateJobStatus
    UpdateJobStatus -->|Next| IncrementUsage
    IncrementUsage -->|Next: End| Success([End: Success])

    ValidateInput -.->|Catch| HandleError
    ParsePrompt -.->|Catch| HandleError
    PlanScenes -.->|Catch| HandleError
    GenerateScenes -.->|Catch| HandleError
    ComposeVideo -.->|Catch| HandleError
    UpdateJobStatus -.->|Catch| HandleError
    IncrementUsage -.->|Catch| HandleError

    HandleError --> Failed([End: Failed])

    style GenerateScenes fill:#e1f5ff,stroke:#0288d1,stroke-width:3px
    style ComposeVideo fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style HandleError fill:#ffcdd2,stroke:#c62828,stroke-width:2px
    style Success fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
```

---

## Parallel Scene Generation (Map State)

Detailed view of the Map state that generates 5 scenes in parallel.

```mermaid
sequenceDiagram
    participant SFN as Step Functions<br/>Map State
    participant L1 as Lambda 1<br/>Scene 1
    participant L2 as Lambda 2<br/>Scene 2
    participant L3 as Lambda 3<br/>Scene 3
    participant L4 as Lambda 4<br/>Scene 4
    participant L5 as Lambda 5<br/>Scene 5
    participant Replicate as Replicate AI
    participant S3 as S3 Assets

    Note over SFN: Input: {scenes: [scene1, scene2, scene3, scene4, scene5]}

    par Parallel Execution (MaxConcurrency: 5)
        SFN->>L1: Invoke Lambda<br/>{sceneIndex: 0, prompt: "opening shot"}
        SFN->>L2: Invoke Lambda<br/>{sceneIndex: 1, prompt: "main action"}
        SFN->>L3: Invoke Lambda<br/>{sceneIndex: 2, prompt: "transition"}
        SFN->>L4: Invoke Lambda<br/>{sceneIndex: 3, prompt: "climax"}
        SFN->>L5: Invoke Lambda<br/>{sceneIndex: 4, prompt: "closing shot"}
    end

    Note over L1,L5: Cold Start: 1-3s (if Lambda not warm)

    par Lambda Execution
        L1->>Replicate: POST /v1/predictions<br/>Model: runway-gen3-turbo
        L2->>Replicate: POST /v1/predictions
        L3->>Replicate: POST /v1/predictions
        L4->>Replicate: POST /v1/predictions
        L5->>Replicate: POST /v1/predictions
    end

    Replicate-->>L1: {"id": "abc1", "status": "processing"}
    Replicate-->>L2: {"id": "abc2", "status": "processing"}
    Replicate-->>L3: {"id": "abc3", "status": "processing"}
    Replicate-->>L4: {"id": "abc4", "status": "processing"}
    Replicate-->>L5: {"id": "abc5", "status": "processing"}

    loop Poll Status (every 5s, max 60s)
        par Polling
            L1->>Replicate: GET /v1/predictions/abc1
            L2->>Replicate: GET /v1/predictions/abc2
            L3->>Replicate: GET /v1/predictions/abc3
            L4->>Replicate: GET /v1/predictions/abc4
            L5->>Replicate: GET /v1/predictions/abc5
        end

        Replicate-->>L1: {"status": "processing"}
        Replicate-->>L2: {"status": "processing"}
        Replicate-->>L3: {"status": "succeeded", "output": "https://..."}
        Replicate-->>L4: {"status": "processing"}
        Replicate-->>L5: {"status": "processing"}

        Note over L3: Scene 3 complete first (fastest generation)
    end

    Note over L1,L5: Scenes complete at different times (30-60s each)

    par Upload to S3
        L1->>S3: PutObject<br/>omnigen-assets/{jobId}/scene-001.mp4
        L2->>S3: PutObject<br/>omnigen-assets/{jobId}/scene-002.mp4
        L3->>S3: PutObject<br/>omnigen-assets/{jobId}/scene-003.mp4
        L4->>S3: PutObject<br/>omnigen-assets/{jobId}/scene-004.mp4
        L5->>S3: PutObject<br/>omnigen-assets/{jobId}/scene-005.mp4
    end

    S3-->>L1: Upload complete (50 MB)
    S3-->>L2: Upload complete (50 MB)
    S3-->>L3: Upload complete (50 MB)
    S3-->>L4: Upload complete (50 MB)
    S3-->>L5: Upload complete (50 MB)

    L1-->>SFN: Return {"sceneUrl": "s3://...scene-001.mp4"}
    L2-->>SFN: Return {"sceneUrl": "s3://...scene-002.mp4"}
    L3-->>SFN: Return {"sceneUrl": "s3://...scene-003.mp4"}
    L4-->>SFN: Return {"sceneUrl": "s3://...scene-004.mp4"}
    L5-->>SFN: Return {"sceneUrl": "s3://...scene-005.mp4"}

    Note over SFN: Map state waits for ALL Lambdas to complete<br/>Total time: max(L1, L2, L3, L4, L5) ≈ 60s
```

**Map State Configuration:**
```json
{
  "Type": "Map",
  "ItemsPath": "$.scenes",
  "MaxConcurrency": 5,
  "Iterator": {
    "StartAt": "GenerateScene",
    "States": {
      "GenerateScene": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:us-east-1:123456789012:function:omnigen-generator",
        "Retry": [
          {
            "ErrorEquals": ["Replicate.APIError"],
            "IntervalSeconds": 5,
            "MaxAttempts": 3,
            "BackoffRate": 2.0
          }
        ],
        "Catch": [
          {
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "HandleSceneError"
          }
        ],
        "End": true
      }
    }
  },
  "ResultPath": "$.sceneResults",
  "Next": "ComposeVideo"
}
```

**Performance Metrics:**
- **Cold Start:** 1-3 seconds per Lambda (first invocation)
- **Warm Start:** <100ms (subsequent invocations within 15 min)
- **Scene Generation:** 30-60 seconds per scene (Replicate API)
- **Total Parallel Time:** ~60 seconds (slowest scene)
- **Sequential Time (if not parallel):** ~300 seconds (5 x 60s)
- **Time Savings:** 80% (300s → 60s)

---

## Error Handling and Retry Logic

Comprehensive error handling strategy with exponential backoff.

```mermaid
flowchart TB
    Invoke([Lambda Invocation])

    Invoke --> Execute{Execute Lambda}

    Execute -->|Success| Return([Return Result])

    Execute -->|Error| ErrorType{Error Type}

    ErrorType -->|Throttling<br/>TooManyRequestsException| Retry1{Retry Attempt 1}
    ErrorType -->|Timeout<br/>Task timed out| Retry1
    ErrorType -->|Replicate API Error<br/>503 Service Unavailable| Retry1
    ErrorType -->|Network Error<br/>Connection timeout| Retry1

    Retry1 -->|Wait 2s| Execute
    Retry1 -->|Max Attempts: 3| Retry2{Retry Attempt 2}

    Retry2 -->|Wait 4s<br/>Backoff Rate: 2.0| Execute
    Retry2 -->|Max Attempts: 3| Retry3{Retry Attempt 3}

    Retry3 -->|Wait 8s<br/>Backoff Rate: 2.0| Execute
    Retry3 -->|All Retries Exhausted| Catch

    ErrorType -->|Fatal Errors| Catch
    subgraph FatalErrors[\"Non-Retryable Errors\"]
        F1[InvalidInputException]
        F2[ResourceNotFoundException]
        F3[UnauthorizedException]
        F4[ValidationException]
    end

    Catch([Catch Error]) --> HandleError[HandleError State]

    HandleError --> LogError[Log to CloudWatch<br/>{<br/>  jobId,<br/>  error,<br/>  stackTrace,<br/>  timestamp<br/>}]

    LogError --> UpdateDDB[Update DynamoDB<br/>Job Status: failed<br/>Error Message: {error}]

    UpdateDDB --> SNS{SNS Notification?}

    SNS -->|Enabled| SendSNS[Send SNS Topic<br/>Subject: Video Generation Failed<br/>Body: {jobId, error}]
    SNS -->|Disabled MVP| Skip

    SendSNS --> End([End: Failed])
    Skip --> End

    style FatalErrors fill:#ffcdd2,stroke:#c62828,stroke-width:2px
    style Retry1 fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style Retry2 fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style Retry3 fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style HandleError fill:#ffcdd2,stroke:#c62828,stroke-width:2px
```

**Retry Configuration:**

| Error Type | Retries | Initial Interval | Backoff Rate | Max Interval |
|------------|---------|------------------|--------------|--------------|
| **Throttling** (Lambda) | 3 | 2s | 2.0 | 8s |
| **Timeout** (Lambda) | 2 | 5s | 1.5 | 7.5s |
| **Replicate API Error** | 3 | 5s | 2.0 | 20s |
| **Network Error** | 3 | 2s | 2.0 | 8s |
| **S3 Upload Error** | 3 | 1s | 2.0 | 4s |
| **DynamoDB Throttle** | 5 | 1s | 2.0 | 16s |
| **Fatal Errors** | 0 | N/A | N/A | N/A |

**Example Retry Sequence:**
```
Attempt 1: Execute → Error (503) → Wait 2s
Attempt 2: Execute → Error (503) → Wait 4s (2s * 2.0)
Attempt 3: Execute → Error (503) → Wait 8s (4s * 2.0)
Attempt 4: Execute → Error (503) → Catch Error → HandleError
```

---

## Timeline Visualization

End-to-end timeline for a 30-second video (5 scenes).

```mermaid
gantt
    title Video Generation Pipeline Timeline (30-second video)
    dateFormat ss
    axisFormat %S s

    section Workflow Start
    ValidateInput           :done, validate, 00, 1s
    ParsePrompt            :done, parse, after validate, 2s
    PlanScenes             :done, plan, after parse, 3s

    section Scene Generation (Parallel)
    Scene 1: Generate      :active, scene1, after plan, 45s
    Scene 2: Generate      :active, scene2, after plan, 50s
    Scene 3: Generate      :active, scene3, after plan, 40s
    Scene 4: Generate      :active, scene4, after plan, 55s
    Scene 5: Generate      :active, scene5, after plan, 48s

    section Video Composition
    Download Scenes        :crit, download, after scene4, 10s
    FFmpeg Processing      :crit, ffmpeg, after download, 60s
    Upload Final Video     :crit, upload, after ffmpeg, 15s

    section Finalization
    Update Job Status      :done, update, after upload, 2s
    Increment Usage        :done, increment, after update, 2s
    Send Notification      :done, notify, after increment, 1s

    section Total Time
    Complete Workflow      :milestone, after notify, 0s
```

**Timeline Breakdown:**

| Phase | Duration | Percentage | Bottleneck |
|-------|----------|------------|------------|
| **Input Validation** | 1s | 0.5% | Minimal |
| **Prompt Parsing** | 2s | 1.1% | LLM API call |
| **Scene Planning** | 3s | 1.6% | LLM API call |
| **Scene Generation** | 55s | 29.4% | Replicate API (slowest scene) |
| **Download Scenes** | 10s | 5.3% | S3 bandwidth |
| **FFmpeg Processing** | 60s | 32.1% | CPU-bound encoding |
| **Upload Final Video** | 15s | 8.0% | S3 bandwidth (200 MB) |
| **Finalization** | 5s | 2.7% | DynamoDB + SNS |
| **TOTAL** | ~187s (3m 7s) | 100% | FFmpeg + Replicate |

**Optimization Opportunities:**
1. **Replicate API:** Use faster models (Runway Gen-3 Turbo vs Standard)
2. **FFmpeg:** Hardware encoding (H.264 NVENC GPU) - requires GPU Lambda (not available)
3. **Download Scenes:** Multipart concurrent downloads (boto3 TransferConfig)
4. **Upload Final:** Multipart upload (already implemented for >5GB files)

---

## State Transitions

Detailed flowchart of all possible state transitions.

```mermaid
flowchart TD
    Start([ECS API calls<br/>StartExecution])

    Start --> S1{ValidateInput}

    S1 -->|Valid| S2[ParsePrompt]
    S1 -->|Invalid| E1[Error: Invalid Input]

    S2 -->|Success| S3[PlanScenes]
    S2 -->|Error| E2[Error: Prompt Parse Failed]

    S3 -->|Success| S4{GenerateScenes<br/>Map State}
    S3 -->|Error| E3[Error: Scene Planning Failed]

    S4 --> S4A[Scene 1]
    S4 --> S4B[Scene 2]
    S4 --> S4C[Scene 3]
    S4 --> S4D[Scene 4]
    S4 --> S4E[Scene 5]

    S4A -->|Success| S4Merge{All Scenes Complete?}
    S4B -->|Success| S4Merge
    S4C -->|Success| S4Merge
    S4D -->|Success| S4Merge
    S4E -->|Success| S4Merge

    S4A -->|Error| E4A[Retry Scene 1<br/>Max 3 attempts]
    S4B -->|Error| E4B[Retry Scene 2<br/>Max 3 attempts]
    S4C -->|Error| E4C[Retry Scene 3<br/>Max 3 attempts]
    S4D -->|Error| E4D[Retry Scene 4<br/>Max 3 attempts]
    S4E -->|Error| E4E[Retry Scene 5<br/>Max 3 attempts]

    E4A -->|Retry Success| S4Merge
    E4B -->|Retry Success| S4Merge
    E4C -->|Retry Success| S4Merge
    E4D -->|Retry Success| S4Merge
    E4E -->|Retry Success| S4Merge

    E4A -->|All Retries Failed| E4[Error: Scene Generation Failed]
    E4B -->|All Retries Failed| E4
    E4C -->|All Retries Failed| E4
    E4D -->|All Retries Failed| E4
    E4E -->|All Retries Failed| E4

    S4Merge -->|Yes| S5[ComposeVideo]

    S5 -->|Success| S6[UpdateJobStatus]
    S5 -->|Error| E5[Error: Composition Failed<br/>Retry 1 attempt]

    E5 -->|Retry Success| S6
    E5 -->|Retry Failed| E5Final[Error: Video Composition Failed]

    S6 -->|Success| S7[IncrementUsage]
    S6 -->|Error| E6[Error: DynamoDB Update Failed<br/>Retry 3 attempts]

    E6 -->|Retry Success| S7
    E6 -->|Retry Failed| E6Final[Error: Job Status Update Failed]

    S7 -->|Success| S8[SendNotification optional]
    S7 -->|Error| E7[Error: Usage Increment Failed]

    S8 -->|Success| End([End: Success<br/>Job Status: completed])
    S8 -->|Error| E8[Error: Notification Failed]

    E1 --> HandleError
    E2 --> HandleError
    E3 --> HandleError
    E4 --> HandleError
    E5Final --> HandleError
    E6Final --> HandleError
    E7 --> HandleError
    E8 --> HandleError

    HandleError[HandleError State] --> LogError[Log Error to CloudWatch]
    LogError --> UpdateFailed[Update Job Status: failed]
    UpdateFailed --> Failed([End: Failed<br/>Job Status: failed])

    style S4 fill:#e1f5ff,stroke:#0288d1,stroke-width:3px
    style S5 fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style HandleError fill:#ffcdd2,stroke:#c62828,stroke-width:2px
    style End fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style Failed fill:#ffcdd2,stroke:#c62828,stroke-width:2px
```

---

## Lambda Function Details

### Generator Lambda (Scene Generation)

**Function Name:** `omnigen-generator`
**Runtime:** Node.js 20.x
**Memory:** 2048 MB
**Timeout:** 900 seconds (15 minutes)
**Ephemeral Storage:** 512 MB (/tmp)
**Concurrency:** 5 (limited by MaxConcurrency in Map state)

**Environment Variables:**
```bash
REPLICATE_SECRET_ARN=arn:aws:secretsmanager:us-east-1:123456789012:secret:omnigen/replicate-api-key
ASSETS_BUCKET=omnigen-assets
REGION=us-east-1
```

**IAM Permissions:**
- `secretsmanager:GetSecretValue` (Replicate API key)
- `s3:PutObject` (upload scenes to S3)
- `dynamodb:UpdateItem` (update job status optional)
- `logs:CreateLogStream`, `logs:PutLogEvents` (CloudWatch Logs)

**Code Structure:**
```javascript
// index.js
const { SecretsManagerClient, GetSecretValueCommand } = require("@aws-sdk/client-secrets-manager");
const { S3Client, PutObjectCommand } = require("@aws-sdk/client-s3");
const Replicate = require("replicate");

exports.handler = async (event) => {
  const { jobId, sceneIndex, prompt, style } = event;

  // 1. Fetch Replicate API key from Secrets Manager (cached)
  const apiKey = await getReplicateKey();
  const replicate = new Replicate({ auth: apiKey });

  // 2. Call Replicate API (Runway Gen-3 Turbo)
  const prediction = await replicate.run(
    "runway-ml/gen-3-turbo",
    {
      input: {
        prompt: prompt,
        duration: 5,  // 5 seconds per scene
        aspect_ratio: "16:9",
        style: style
      }
    }
  );

  // 3. Poll for completion (max 60s)
  let output = prediction.output;
  while (!output && prediction.status !== "failed") {
    await new Promise(resolve => setTimeout(resolve, 5000)); // 5s interval
    prediction = await replicate.predictions.get(prediction.id);
    output = prediction.output;
  }

  if (prediction.status === "failed") {
    throw new Error(`Replicate generation failed: ${prediction.error}`);
  }

  // 4. Download video from Replicate
  const response = await fetch(output);
  const buffer = await response.arrayBuffer();

  // 5. Upload to S3
  const s3 = new S3Client({ region: process.env.REGION });
  await s3.send(new PutObjectCommand({
    Bucket: process.env.ASSETS_BUCKET,
    Key: `${jobId}/scene-${String(sceneIndex + 1).padStart(3, '0')}.mp4`,
    Body: Buffer.from(buffer),
    ContentType: 'video/mp4'
  }));

  return {
    sceneIndex,
    sceneUrl: `s3://${process.env.ASSETS_BUCKET}/${jobId}/scene-${String(sceneIndex + 1).padStart(3, '0')}.mp4`,
    duration: 5,
    status: "success"
  };
};
```

**Cost per Invocation:**
- Lambda execution: $0.002 (2048 MB x 60s)
- Replicate API: $0.26 (Runway Gen-3 Turbo, 5s video)
- S3 PUT request: $0.000005
- **Total: $0.26/scene**

### Composer Lambda (Video Stitching)

**Function Name:** `omnigen-composer`
**Runtime:** Node.js 20.x
**Memory:** 10240 MB (10 GB for FFmpeg)
**Timeout:** 900 seconds (15 minutes)
**Ephemeral Storage:** 10240 MB (10 GB for /tmp)
**Concurrency:** 1 (sequential processing)

**Environment Variables:**
```bash
ASSETS_BUCKET=omnigen-assets
REGION=us-east-1
```

**Lambda Layer:**
- `ffmpeg-layer`: FFmpeg 6.0 static binary (50 MB)

**IAM Permissions:**
- `s3:GetObject` (download scenes)
- `s3:PutObject` (upload final video)
- `dynamodb:UpdateItem` (update job status)
- `logs:CreateLogStream`, `logs:PutLogEvents`

**Code Structure:**
```javascript
// index.js
const { S3Client, GetObjectCommand, PutObjectCommand } = require("@aws-sdk/client-s3");
const { spawn } = require("child_process");
const fs = require("fs");
const path = require("path");

exports.handler = async (event) => {
  const { jobId, sceneResults } = event;

  // 1. Download all scenes to /tmp
  const s3 = new S3Client({ region: process.env.REGION });
  for (const scene of sceneResults) {
    const { sceneUrl } = scene;
    const key = sceneUrl.replace(`s3://${process.env.ASSETS_BUCKET}/`, '');
    const response = await s3.send(new GetObjectCommand({
      Bucket: process.env.ASSETS_BUCKET,
      Key: key
    }));
    const buffer = await streamToBuffer(response.Body);
    fs.writeFileSync(`/tmp/scene-${scene.sceneIndex + 1}.mp4`, buffer);
  }

  // 2. Create concat file for FFmpeg
  const concatFile = sceneResults.map((s, i) =>
    `file '/tmp/scene-${i + 1}.mp4'`
  ).join('\n');
  fs.writeFileSync('/tmp/concat.txt', concatFile);

  // 3. Run FFmpeg to concatenate with transitions
  await runFFmpeg([
    '-f', 'concat',
    '-safe', '0',
    '-i', '/tmp/concat.txt',
    '-vf', 'fade=t=in:st=0:d=0.5,fade=t=out:st=29.5:d=0.5,scale=1920:1080',
    '-c:v', 'libx264',
    '-preset', 'fast',
    '-crf', '23',
    '-c:a', 'aac',
    '-b:a', '192k',
    '/tmp/final.mp4'
  ]);

  // 4. Upload final video to S3
  const finalBuffer = fs.readFileSync('/tmp/final.mp4');
  await s3.send(new PutObjectCommand({
    Bucket: process.env.ASSETS_BUCKET,
    Key: `${jobId}/final.mp4`,
    Body: finalBuffer,
    ContentType: 'video/mp4'
  }));

  return {
    videoUrl: `s3://${process.env.ASSETS_BUCKET}/${jobId}/final.mp4`,
    duration: sceneResults.length * 5,
    status: "success"
  };
};

function runFFmpeg(args) {
  return new Promise((resolve, reject) => {
    const ffmpeg = spawn('/opt/bin/ffmpeg', args);
    ffmpeg.on('close', code => {
      if (code === 0) resolve();
      else reject(new Error(`FFmpeg exited with code ${code}`));
    });
  });
}
```

**Cost per Invocation:**
- Lambda execution: $0.01 (10240 MB x 90s)
- S3 GET requests: 5 x $0.0000004 = $0.000002
- S3 PUT request: $0.000005
- **Total: $0.01/video**

---

## Monitoring and Observability

### CloudWatch Metrics

**Step Functions Metrics:**
- `ExecutionTime` (ms) - Track workflow duration
- `ExecutionsFailed` (count) - Track failure rate
- `ExecutionsSucceeded` (count) - Track success rate
- `ExecutionsTimedOut` (count) - Track timeouts (should be 0)

**Lambda Metrics:**
- `Duration` (ms) - Generator: 30-60s, Composer: 60-90s
- `Errors` (count) - Track Lambda errors
- `Throttles` (count) - Track concurrency throttling
- `ConcurrentExecutions` (count) - Track parallelism

**Custom Metrics (Future):**
- `SceneGenerationTime` (ms) - Per-scene generation time
- `FFmpegEncodingTime` (ms) - Video composition time
- `ReplicateAPILatency` (ms) - External API latency

### CloudWatch Logs

**Log Groups:**
- `/aws/lambda/omnigen-generator` - Scene generation logs
- `/aws/lambda/omnigen-composer` - Video composition logs
- `/aws/states/omnigen-workflow` - Step Functions execution logs

**Log Insights Queries:**

**1. Average Video Generation Time:**
```sql
fields @timestamp, @message
| filter @message like /Execution succeeded/
| stats avg(@duration) as avg_duration, max(@duration) as max_duration
```

**2. Failed Scenes:**
```sql
fields @timestamp, jobId, sceneIndex, error
| filter @message like /Replicate generation failed/
| stats count() by error
```

**3. FFmpeg Encoding Errors:**
```sql
fields @timestamp, jobId, @message
| filter @message like /FFmpeg exited with code/
| display jobId, @message
```

---

## Cost Analysis

### Per-Video Cost (30-second video, 5 scenes)

| Component | Cost | Notes |
|-----------|------|-------|
| **Step Functions Express** | $0.000025 | $1.00 per million state transitions |
| **Generator Lambda (5x)** | $0.01 | 5 invocations x $0.002 each |
| **Composer Lambda (1x)** | $0.01 | 1 invocation x $0.01 |
| **DynamoDB Write Requests** | $0.0000125 | 10 writes x $0.00000125 |
| **S3 Storage (1 month)** | $0.005 | 200 MB x $0.023/GB |
| **S3 PUT Requests** | $0.00003 | 6 PUTs x $0.000005 |
| **Replicate API (5x)** | $1.30 | 5 scenes x $0.26 (Runway Gen-3 Turbo) |
| **TOTAL** | **$1.32** | Well under $2.00 target |

### Monthly Cost (100 videos)

| Component | Cost |
|-----------|------|
| **Infrastructure (idle)** | $100.00 |
| **Video Generation (100x)** | $132.00 |
| **Data Transfer (CloudFront)** | $8.50 |
| **TOTAL** | **$240.50** |

---

**Related Documentation:**
- [Architecture Overview](./architecture-overview.md) - System design
- [Data Flow](./data-flow.md) - Complete request/response flows
- [Backend Architecture](./backend-architecture.md) - Go API triggering workflows
