# Learnyst - Enterprise Multi-Tenant EdTech SaaS Platform

Learnyst is a production-grade, highly scalable Learning Management System (LMS) designed for course creators, universities, and coaching academies to host, manage, and securely deliver online courses. Built with a modern microservices architecture in Go and Next.js, it offers robust multi-tenancy, a custom event-driven video transcoding pipeline, and a stunning, high-contrast brutalist UI ("Ada Aesthetic").

## 🚀 Key Features

*   **True Multi-Tenancy:** Each academy operates as an isolated tenant with custom subdomains (e.g., `academy1.learnyst.com`), segregated data, and dedicated branding.
*   **Adaptive HLS Video Streaming:** 
    *   Lightning-fast local `FFMPEG` processing via shared Docker volumes (bypasses slow cloud egress).
    *   Automated multi-bitrate HLS chunking (1080p, 720p, 480p) optimized for zero-buffering with strict GOP forcing (`-g 48 -hls_time 4`).
    *   Direct-to-cloud sync to Azure Blob Storage with automated CORS configuration.
    *   Fallback mechanics to handle videos lacking audio streams.
*   **Asynchronous Event-Driven Pipeline:** Leverages Kafka message brokering to decouple heavy workloads like video transcoding from core API traffic, ensuring 99.9% uptime.
*   **Ada Aesthetic UI:** A bold, high-contrast, brutalist-inspired Next.js frontend built with Tailwind CSS, delivering a highly premium and modern learning experience.
*   **Comprehensive Course Management:** Support for dynamic curriculum building, including video lectures, downloadable lecture notes (PDFs), and assignments.
*   **Role-Based Access Control (RBAC):** Distinct portals and capabilities for System Admins, Tenant Owners, Teachers, and Students.

---

## 🏗️ Architecture & Microservices

Learnyst uses a highly decoupled microservices architecture to ensure scalability, fault tolerance, and developer velocity.

### 1. `lms-api-gateway`
*   **Role:** The central entry point for all frontend client requests.
*   **Stack:** Go (Gin).
*   **Responsibilities:** Handles global rate limiting, CORS configuration, JWT authentication verification, and routes authenticated traffic to the respective downstream microservices.

### 2. `lms-core-service`
*   **Role:** The brains of the platform.
*   **Stack:** Go (Gin, GORM).
*   **Responsibilities:** Manages tenant creation, course logic, user enrollments, staff assignments, and handles the initial video ingestion. It intercepts raw `.mp4` uploads and saves them to a shared local Docker volume (`/shared`), preventing expensive and slow double-hops to cloud storage.

### 3. `lms-video-worker`
*   **Role:** Heavy-duty asynchronous processor.
*   **Stack:** Go, Sarama (Kafka), FFMPEG.
*   **Responsibilities:** Consumes `video.uploaded` Kafka events. Reads the raw video from the local shared disk, spins up an optimized FFMPEG sub-process to generate keyframe-aligned multi-bitrate `.m3u8` playlists and `.ts` chunks, pushes them to Azure Blob Storage, and emits a `video.transcoded` event upon completion.

### 4. `tenant-proxy-service`
*   **Role:** Subdomain routing layer.
*   **Stack:** Go.
*   **Responsibilities:** Intercepts traffic at the edge, reads the HTTP Host header, resolves the subdomain to a specific Tenant ID in the database, and injects tenant context into downstream requests.

### 5. `lms-frontend`
*   **Role:** User-facing web application.
*   **Stack:** Next.js 14, React, Tailwind CSS, TypeScript, Video.js.
*   **Responsibilities:** Provides the "Ada Aesthetic" dashboards for teachers to build courses and students to consume content via custom Video.js HLS players.

---

## 🗄️ Database & Storage Layer

*   **PostgreSQL 17:** Primary relational datastore (Users, Courses, Tenants, Roles). Configured with replica instances for read-heavy operations.
*   **Redis 7:** High-speed in-memory datastore for caching tenant lookups, rate limiting, and session management.
*   **Kafka 7.5 & Zookeeper:** High-throughput message broker handling the event stream for background jobs (e.g., `video.uploaded`, `video.transcoded`, `email.send`).
*   **Cassandra 4.1:** High-throughput NoSQL storage designed for writing massive amounts of analytical data and tracking user video progression/logs.
*   **Azure Blob Storage:** Long-term cold and warm storage for serving public/private HLS video chunks, PDF notes, and image thumbnails.

---

## 📦 Deep Dive: The Video Pipeline

Our optimized HLS pipeline is built to solve the notorious "infinite buffering" and "cloud egress" problems typical in VOD platforms:
1.  **Ingestion:** The Core Service receives an `.mp4` and saves it instantly to a shared local Docker volume (`/shared`).
2.  **Event Publishing:** A `video.uploaded` Kafka event is emitted containing the local `local:///shared/filename.mp4` path.
3.  **Transcoding:** The Video Worker consumes the event, grabs the file instantly from disk, and runs an optimized FFMPEG command:
    *   `-preset veryfast`: Ensures milliseconds processing time.
    *   `-g 48 -keyint_min 48 -sc_threshold 0`: Forces keyframes every 2 seconds, which is absolutely critical for preventing HLS players from stalling or endlessly buffering.
    *   `-hls_time 4`: Creates small, 4-second chunks so the browser starts playing almost instantly.
4.  **Cloud Sync:** The finalized multi-bitrate HLS streams are uploaded to Azure Blob Storage (which has CORS programmatically configured to allow frontend fetches).
5.  **Cleanup & Notification:** The local `.mp4` is deleted to conserve server disk space, and a `video.transcoded` Kafka event is sent back to the Core Service to update the database with the public Azure `.m3u8` URL.

---

## 🛠️ Getting Started & Local Development

### Prerequisites
*   Docker & Docker Compose (V2+)
*   Node.js (v18+)
*   Go (1.21+)
*   Azure Blob Storage Account credentials (or Azurite Emulator). Ensure CORS is configured to allow `GET` and `OPTIONS` from `*` or `localhost:3005`.

### 1. Bootstrapping the Backend Infrastructure

The entire backend infrastructure, including databases and all Go microservices, is orchestrated via Docker Compose.

```bash
cd lms-api-gateway
# Ensure the .env file exists with valid POSTGRES and AZURE credentials
docker-compose up -d --build
```
*You can run `docker-compose logs -f lms-video-worker` to watch the FFMPEG pipeline in action when you upload a video.*

### 2. Bootstrapping the Frontend

```bash
cd lms-frontend
npm install
npm run dev
```
The Next.js application will compile and be accessible at `http://localhost:3005`.

---

## 🤝 Contributing & Standards

*   **Go Code:** Ensure all Go code is formatted using `gofmt` and follows idiomatic Go practices. Avoid unnecessary comments unless documenting complex business logic.
*   **Frontend Code:** Strictly adhere to the "Ada Aesthetic". Utilize Tailwind utility classes, maintain high contrast ratios, and use brutalist design principles (sharp corners, solid borders).

## 📄 License

This project is proprietary and confidential. All rights reserved.
