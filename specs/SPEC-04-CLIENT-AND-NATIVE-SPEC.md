# SPEC-04: PulseOps AI — Web, Mobile & Native Platform Architecture

## 1. Web Architecture (Next.js 16.2.6 & React 19.2)

### 1.1 Rendering & Performance Strategy
- **Partial Prerendering (PPR):** Core layout, navigation, and static incident historical views prerendered statically at build time.
- **Server Actions v2:** Mutation execution (`triageIncident`, `drainWorker`) using zero-bundle-cost Server Actions with immediate optimistic cache updates.
- **Virtualization:** Incident logs and GPU telemetry stream rendered with `@tanstack/react-virtual` to maintain constant 60fps even with 50,000+ streaming lines.
- **Client State vs Server State:**
  - *Server State:* Managed by Apollo Client 3.12 normalized cache.
  - *Local Canvas State:* Managed by **Zustand 5.x** (for interactive GPU cluster topology nodes).

### 1.2 Typed GraphQL Code Generation (`@graphql-codegen` v5)
`codegen.ts` generates strongly typed hooks directly from `SPEC-03-GRAPHQL-SCHEMA.graphql`:
```typescript
import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../packages/api-graphql/schema.graphql',
  documents: ['src/**/*.graphql'],
  generates: {
    './src/generated/graphql.ts': {
      plugins: ['typescript', 'typescript-operations', 'typescript-react-apollo'],
      config: {
        withHooks: true,
        useIndexSignature: true,
      },
    },
  },
};
export default config;
```

---

## 2. Mobile Architecture (Expo SDK 56 & React Native 0.85.3)

### 2.1 File-Based Routing (`expo-router` v56)
- `app/(tabs)/index.tsx`: Live incident dashboard with pull-to-refresh and optimistic triage swipe actions.
- `app/(tabs)/fleet.tsx`: GPU cluster health grid powered by **Shopify FlashList** for smooth scrolling.
- `app/incident/[id].tsx`: Incident war room with audio/call bridge and auto-heal execution logs.

### 2.2 Offline First & Resilience
- Local caching with **Expo SQLite (OPFS driver)**.
- Offline actions (e.g. on-call engineer pressing "Acknowledge" while on an airplane or subway) are stored in a local SQLite queue and replayed idempotently upon network reconnect.

---

## 3. Native Platform Modules (Swift / Kotlin)

To fulfill the JD requirement:
> *"Apply a basic understanding of native mobile development concepts in Kotlin/Java and Objective-C/Swift when working with React Native integrations, debugging, or platform-specific requirements"*

PulseOps AI implements a custom native Expo module: `modules/native-incident-alert`.

### 3.1 iOS Native Module (Swift 6 / iOS 18)
- **Live Activity & Dynamic Island Integration:** Employs `ActivityKit` to project incident duration, affected GPU count, and auto-heal status onto the Lock Screen and Dynamic Island.
- **Critical Alerts:** Configures `UNAuthorizationOptions.criticalAlert` so severity-critical GPU failures trigger an audible siren even when the phone is on mute or "Do Not Disturb".

### 3.2 Android Native Module (Kotlin 2.x / Android 15)
- **Foreground Service with Full-Screen Intent:**
  When a Critical Incident arrives, the app triggers a high-priority heads-up notification channel:
  ```kotlin
  val channel = NotificationChannel(
      "INCIDENT_CRITICAL",
      "Critical GPU Incidents",
      NotificationManager.IMPORTANCE_HIGH
  ).apply {
      setBypassDnd(true)
      enableVibration(true)
  }
  ```
