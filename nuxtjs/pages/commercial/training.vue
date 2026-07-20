<template>
  <div data-testid="commercial-training-page">
    <ProPageHeader
      :title="$t('training.title')"
      :subtitle="$t('training.subtitle')"
    />

    <div class="pf-training-tabs pro-mb-lg">
      <ProButton
        :variant="tab === 'call' ? 'primary' : 'ghost'"
        test-id="training-tab-call"
        @click="tab = 'call'"
      >
        {{ $t('training.tabCall') }}
      </ProButton>
      <ProButton
        :variant="tab === 'history' ? 'primary' : 'ghost'"
        test-id="training-tab-history"
        @click="tab = 'history'; loadHistory()"
      >
        {{ $t('training.tabHistory') }}
      </ProButton>
      <ProButton
        :variant="tab === 'scripts' ? 'primary' : 'ghost'"
        test-id="training-tab-scripts"
        @click="tab = 'scripts'"
      >
        {{ $t('training.tabScripts') }}
      </ProButton>
    </div>

    <template v-if="tab === 'call'">
      <div class="pf-training-grid">
        <ProCard :title="$t('training.scriptFollow')" class="pf-training-script">
          <label class="pro-label">{{ $t('training.selectScript') }}</label>
          <select v-model="scriptId" class="pro-select" data-testid="training-script-select" :disabled="phase !== 'idle'">
            <option v-for="s in scripts" :key="s.id" :value="s.id">
              {{ s.title }}{{ s.ownerUserId ? ` (${$t('training.personalized')})` : '' }}
            </option>
          </select>
          <ol v-if="selectedScript" class="pf-steps">
            <li v-for="step in selectedSteps" :key="step.id || step.title">
              <strong>{{ step.title }}</strong>
              <ul>
                <li v-for="p in step.talkingPoints || []" :key="p">{{ p }}</li>
              </ul>
              <p v-if="step.exampleLine" class="pro-hint">« {{ step.exampleLine }} »</p>
            </li>
          </ol>
          <details v-if="selectedDialogue.length" class="pf-example">
            <summary>{{ $t('training.exampleDialogue') }}</summary>
            <p v-for="(line, i) in selectedDialogue" :key="i" class="pf-dialogue-line">
              <strong>{{ line.role === 'vet' ? $t('training.roleVet') : $t('training.roleCommercial') }} :</strong>
              {{ line.text }}
            </p>
          </details>
        </ProCard>

        <ProCard :title="$t('training.phone')" class="pf-training-phone" data-testid="training-phone">
          <template v-if="phase === 'idle'">
            <p class="pro-hint">{{ $t('training.difficultyHint') }}</p>
            <div class="pf-difficulty" data-testid="training-difficulty">
              <button
                v-for="d in difficulties"
                :key="d.value"
                type="button"
                class="pf-diff-card"
                :class="{ 'pf-diff-card--active': interestLevel === d.value }"
                @click="interestLevel = d.value"
              >
                <strong>{{ $t(`training.difficulty.${d.value}.label`) }}</strong>
                <span>{{ $t(`training.difficulty.${d.value}.desc`) }}</span>
              </button>
            </div>
            <label class="pro-label">{{ $t('training.voice') }}</label>
            <select v-model="voiceName" class="pro-select" data-testid="training-voice">
              <option v-for="v in voices" :key="v" :value="v">{{ v }}</option>
            </select>
            <ProButton
              class="pro-mt-md"
              test-id="training-call-btn"
              :disabled="!scriptId || calling"
              @click="startCall"
            >
              <ProIcon name="phone_in_talk" :size="20" />
              {{ $t('training.callBtn') }}
            </ProButton>
            <p v-if="error" class="pf-error">{{ error }}</p>
          </template>

          <template v-else-if="phase === 'ringing'">
            <div class="pf-ringing" data-testid="training-ringing">
              <ProIcon name="ring_volume" :size="48" />
              <p>{{ $t('training.ringing') }}</p>
            </div>
          </template>

          <template v-else-if="phase === 'in_call'">
            <div class="pf-call-bar">
              <ProBadge>{{ $t(`training.difficulty.${interestLevel}.label`) }}</ProBadge>
              <strong class="pf-countdown" data-testid="training-countdown">{{ formatTime(remainingSec) }}</strong>
              <ProButton variant="ghost" test-id="training-hangup" @click="hangUp('manual')">
                {{ $t('training.hangUp') }}
              </ProButton>
            </div>
            <div class="pf-transcript" data-testid="training-transcript">
              <p v-for="(line, i) in transcript" :key="i" :class="'pf-line pf-line--' + line.role">
                <strong>{{ line.role === 'vet' ? $t('training.roleVet') : $t('training.roleCommercial') }}</strong>
                {{ line.text }}
              </p>
            </div>
            <div class="pf-mic-row">
              <ProButton
                :variant="listening ? 'primary' : 'secondary'"
                test-id="training-mic"
                :disabled="busyTurn"
                @click="toggleListen"
              >
                <ProIcon :name="listening ? 'mic' : 'mic_off'" :size="20" />
                {{ listening ? $t('training.listening') : $t('training.pushToTalk') }}
              </ProButton>
              <form class="pf-text-turn" @submit.prevent="sendTextTurn">
                <input
                  v-model="typedLine"
                  class="pro-input"
                  :placeholder="$t('training.typePlaceholder')"
                  data-testid="training-text-input"
                  :disabled="busyTurn"
                >
                <ProButton type="submit" :disabled="busyTurn || !typedLine.trim()" test-id="training-send">
                  {{ $t('training.send') }}
                </ProButton>
              </form>
            </div>
          </template>

          <template v-else-if="phase === 'analyzing'">
            <p data-testid="training-analyzing">{{ $t('training.analyzing') }}</p>
          </template>

          <template v-else-if="phase === 'done' && sim">
            <div data-testid="training-coach">
              <ProBadge>{{ outcomeLabel }}</ProBadge>
              <p v-if="sim.appointmentSlot" class="pro-hint">{{ sim.appointmentSlot }}</p>
              <h3>{{ $t('training.coachTitle') }} — {{ coachScore }}/10</h3>
              <ul v-if="coachTips.length">
                <li v-for="tip in coachTips" :key="tip">{{ tip }}</li>
              </ul>
              <div v-if="sim.audioUrl || localAudioUrl" class="pro-mt-md">
                <label class="pro-label">{{ $t('training.replay') }}</label>
                <audio controls :src="sim.audioUrl || localAudioUrl" data-testid="training-replay" class="pf-audio" />
              </div>
              <label class="pro-label">{{ $t('training.yourScore') }}</label>
              <input v-model.number="userScore" type="number" min="0" max="10" step="0.5" class="pro-input" data-testid="training-user-score">
              <ProButton variant="secondary" class="pro-mt-md" test-id="training-save-score" @click="saveUserScore">
                {{ $t('training.saveScore') }}
              </ProButton>

              <h3 class="pro-mt-md">{{ $t('training.feedbackTitle') }}</h3>
              <p class="pro-hint">{{ $t('training.feedbackHint') }}</p>
              <label class="pro-label">{{ $t('training.vetRealism') }} (1–5)</label>
              <input v-model.number="fb.vetRealism" type="number" min="1" max="5" class="pro-input">
              <label class="pro-label">{{ $t('training.coachUsefulness') }} (1–5)</label>
              <input v-model.number="fb.coachUsefulness" type="number" min="1" max="5" class="pro-input">
              <label class="pro-label">{{ $t('training.difficultyFelt') }}</label>
              <select v-model="fb.difficultyFelt" class="pro-select">
                <option value="too_easy">{{ $t('training.felt.too_easy') }}</option>
                <option value="ok">{{ $t('training.felt.ok') }}</option>
                <option value="too_hard">{{ $t('training.felt.too_hard') }}</option>
              </select>
              <label class="pro-label">{{ $t('training.comment') }}</label>
              <textarea v-model="fb.comment" class="pro-input" rows="3" data-testid="training-feedback-comment" />
              <div class="pf-fb-actions pro-mt-md">
                <ProButton test-id="training-submit-feedback" @click="submitFeedback(false)">
                  {{ $t('training.submitFeedback') }}
                </ProButton>
                <ProButton
                  v-if="canSkip"
                  variant="ghost"
                  test-id="training-skip-feedback"
                  @click="submitFeedback(true)"
                >
                  {{ $t('training.skipFeedback') }}
                </ProButton>
              </div>
              <ProButton v-if="feedbackDone" class="pro-mt-md" test-id="training-new-call" @click="resetCall">
                {{ $t('training.newCall') }}
              </ProButton>
            </div>
          </template>
        </ProCard>
      </div>
    </template>

    <template v-else-if="tab === 'history'">
      <ProCard :title="$t('training.historyTitle')">
        <table v-if="history.length" class="pro-table">
          <thead>
            <tr>
              <th>{{ $t('training.colDate') }}</th>
              <th>{{ $t('training.colDifficulty') }}</th>
              <th>{{ $t('training.colOutcome') }}</th>
              <th>{{ $t('training.colScore') }}</th>
              <th>Top5</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in history" :key="h.id">
              <td>{{ new Date(h.createdAt).toLocaleString() }}</td>
              <td>{{ $t(`training.difficulty.${h.interestLevel}.label`) }}</td>
              <td>{{ $t(`training.outcome.${h.outcome}`) }}</td>
              <td>{{ h.userScore ?? h.aiScore ?? '—' }}</td>
              <td>
                <span v-if="h.isTop5">★</span>
                <audio v-if="h.audioUrl" controls :src="h.audioUrl" class="pf-audio-mini" />
              </td>
            </tr>
          </tbody>
        </table>
        <ProEmptyState v-else :title="$t('training.historyEmpty')" />
      </ProCard>
    </template>

    <template v-else>
      <ProCard :title="$t('training.scriptsTitle')">
        <div v-for="s in scripts" :key="s.id" class="pf-script-row">
          <div>
            <strong>{{ s.title }}</strong>
            <span class="text-muted"> — {{ s.slug }}</span>
            <ProBadge v-if="s.ownerUserId" variant="success">{{ $t('training.personalized') }}</ProBadge>
          </div>
          <ProButton
            v-if="!s.ownerUserId"
            variant="secondary"
            @click="personalize(s.id)"
          >
            {{ $t('training.personalize') }}
          </ProButton>
        </div>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()

type Script = {
  id: string
  title: string
  slug: string
  ownerUserId?: string
  steps?: any
  exampleDialogue?: any
}
type Line = { role: string, text: string }

const tab = ref<'call' | 'history' | 'scripts'>('call')
const scripts = ref<Script[]>([])
const scriptId = ref('')
const interestLevel = ref('neutre')
const VOICE_KEY = 'pf_training_voice'
const voices = ['Charon', 'Kore', 'Puck', 'Aoede', 'Fenrir', 'Sulafat', 'Orus', 'Leda'] as const
/** Distinct TTS profiles per named voice (browser SpeechSynthesis). */
const voiceProfiles: Record<string, { pitch: number, rate: number, voiceIndex: number }> = {
  Charon: { pitch: 0.85, rate: 0.95, voiceIndex: 0 },
  Kore: { pitch: 1.15, rate: 1.0, voiceIndex: 1 },
  Puck: { pitch: 1.25, rate: 1.1, voiceIndex: 2 },
  Aoede: { pitch: 1.05, rate: 0.98, voiceIndex: 1 },
  Fenrir: { pitch: 0.7, rate: 0.92, voiceIndex: 0 },
  Sulafat: { pitch: 1.1, rate: 0.96, voiceIndex: 2 },
  Orus: { pitch: 0.9, rate: 1.0, voiceIndex: 0 },
  Leda: { pitch: 1.2, rate: 1.05, voiceIndex: 1 },
}
const voiceName = ref(
  (typeof localStorage !== 'undefined' && localStorage.getItem(VOICE_KEY)) || 'Charon',
)
const difficulties = [
  { value: 'hostile' },
  { value: 'sceptique' },
  { value: 'neutre' },
  { value: 'interesse' },
  { value: 'chaud' },
]

const phase = ref<'idle' | 'ringing' | 'in_call' | 'analyzing' | 'done'>('idle')
const calling = ref(false)
const busyTurn = ref(false)
const listening = ref(false)
const error = ref('')
const simId = ref('')
const sim = ref<any>(null)
const transcript = ref<Line[]>([])
const typedLine = ref('')
const remainingSec = ref(8 * 60)
const feedbackDone = ref(false)
const canSkip = ref(true)
const userScore = ref<number | null>(null)
const localAudioUrl = ref('')
const fb = reactive({
  vetRealism: 4,
  coachUsefulness: 4,
  difficultyFelt: 'ok',
  comment: '',
})

let timer: ReturnType<typeof setInterval> | null = null
let recognition: any = null
let mediaRecorder: MediaRecorder | null = null
let recordChunks: Blob[] = []
let recordStream: MediaStream | null = null
const history = ref<any[]>([])

watch(voiceName, (v) => {
  try { localStorage.setItem(VOICE_KEY, v) } catch { /* ignore */ }
})

const selectedScript = computed(() => scripts.value.find(s => s.id === scriptId.value))
const selectedSteps = computed(() => {
  const raw = selectedScript.value?.steps
  return Array.isArray(raw) ? raw : []
})
const selectedDialogue = computed(() => {
  const raw = selectedScript.value?.exampleDialogue
  return Array.isArray(raw) ? raw : []
})

const coachScore = computed(() => sim.value?.coachFeedback?.score ?? sim.value?.aiScore ?? '—')
const coachTips = computed(() => {
  const c = sim.value?.coachFeedback
  if (!c) return []
  return [...(c.coachingTips || []), ...(c.improvements || [])].slice(0, 6)
})
const outcomeLabel = computed(() => t(`training.outcome.${sim.value?.outcome || 'manual'}`))

function formatTime(sec: number) {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${String(s).padStart(2, '0')}`
}

function speak(text: string) {
  if (typeof window === 'undefined' || !window.speechSynthesis) return
  const profile = voiceProfiles[voiceName.value] || voiceProfiles.Charon
  const u = new SpeechSynthesisUtterance(text)
  u.lang = 'fr-FR'
  u.pitch = profile.pitch
  u.rate = profile.rate
  const voicesList = window.speechSynthesis.getVoices().filter(v => v.lang.startsWith('fr'))
  if (voicesList.length) {
    u.voice = voicesList[profile.voiceIndex % voicesList.length]
  }
  window.speechSynthesis.cancel()
  window.speechSynthesis.speak(u)
}

async function startRecording() {
  await stopRecordingAsync(false)
  recordChunks = []
  try {
    recordStream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const mime = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
      ? 'audio/webm;codecs=opus'
      : 'audio/webm'
    mediaRecorder = new MediaRecorder(recordStream, { mimeType: mime })
    mediaRecorder.ondataavailable = (e) => {
      if (e.data?.size) recordChunks.push(e.data)
    }
    mediaRecorder.start(1000)
  } catch {
    // mic denied — text mode still works
  }
}

async function stopRecordingAsync(keepBlob: boolean): Promise<Blob | null> {
  return new Promise((resolve) => {
    const finish = () => {
      if (recordStream) {
        recordStream.getTracks().forEach(t => t.stop())
        recordStream = null
      }
      if (!keepBlob || !recordChunks.length) {
        resolve(null)
        return
      }
      resolve(new Blob(recordChunks, { type: 'audio/webm' }))
    }
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.onstop = () => finish()
      mediaRecorder.stop()
      mediaRecorder = null
    } else {
      mediaRecorder = null
      finish()
    }
  })
}

async function uploadRecording(blob: Blob | null) {
  if (!blob || !simId.value) return
  localAudioUrl.value = URL.createObjectURL(blob)
  const fd = new FormData()
  fd.append('file', blob, 'call.webm')
  try {
    const res: any = await $fetch(`/api/commercial/pitch-sims/${simId.value}/audio`, {
      method: 'POST',
      body: fd,
    })
    const data = res.data ?? res
    if (data?.url) {
      if (sim.value) sim.value.audioUrl = data.url
      else sim.value = { audioUrl: data.url }
    }
  } catch {
    // keep local blob URL for replay
  }
}

async function loadScripts() {
  const res: any = await $fetch('/api/commercial/pitch-scripts')
  scripts.value = res.data ?? res
  if (!scriptId.value && scripts.value.length) scriptId.value = scripts.value[0].id
}

async function loadHistory() {
  const res: any = await $fetch('/api/commercial/pitch-sims')
  history.value = res.data ?? res
}

async function loadSkipQuota() {
  try {
    const res: any = await $fetch('/api/commercial/pitch-sims/skip-quota')
    const data = res.data ?? res
    canSkip.value = !!data.canSkip
  } catch {
    canSkip.value = true
  }
}

async function personalize(id: string) {
  await $fetch(`/api/commercial/pitch-scripts/${id}/personalize`, { method: 'POST' })
  await loadScripts()
}

async function startCall() {
  error.value = ''
  calling.value = true
  phase.value = 'ringing'
  try {
    const res: any = await $fetch('/api/commercial/pitch-sims', {
      method: 'POST',
      body: {
        scriptId: scriptId.value,
        interestLevel: interestLevel.value,
        voiceName: voiceName.value,
      },
    })
    const data = res.data ?? res
    simId.value = data.simulation.id
    transcript.value = [{ role: 'vet', text: data.vetOpening || 'Allo ?' }]
    await new Promise(r => setTimeout(r, 2500))
    phase.value = 'in_call'
    remainingSec.value = data.maxSeconds || 480
    await startRecording()
    speak(data.vetOpening || 'Allo ?')
    startTimer()
  } catch (e: any) {
    error.value = e?.data?.statusMessage || e?.message || t('training.errorStart')
    phase.value = 'idle'
  } finally {
    calling.value = false
  }
}

function startTimer() {
  stopTimer()
  timer = setInterval(() => {
    remainingSec.value -= 1
    if (remainingSec.value <= 0) {
      hangUp('timeout')
    }
  }, 1000)
}

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  if (recognition) {
    try { recognition.stop() } catch { /* ignore */ }
    recognition = null
  }
  listening.value = false
}

async function sendTurn(text: string) {
  if (!text.trim() || busyTurn.value || !simId.value) return
  busyTurn.value = true
  transcript.value.push({ role: 'commercial', text: text.trim() })
  try {
    const res: any = await $fetch(`/api/commercial/pitch-sims/${simId.value}/turn`, {
      method: 'POST',
      body: { text: text.trim() },
    })
    const data = res.data ?? res
    transcript.value = data.transcript || [
      ...transcript.value,
      { role: 'vet', text: data.reply },
    ]
    speak(data.reply)
    if (data.ended) {
      await finalizeCall(data.outcome)
    }
  } catch (e: any) {
    error.value = e?.data?.statusMessage || t('training.errorTurn')
  } finally {
    busyTurn.value = false
  }
}

function sendTextTurn() {
  const line = typedLine.value
  typedLine.value = ''
  return sendTurn(line)
}

function toggleListen() {
  const SR = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
  if (!SR) {
    error.value = t('training.noSpeechApi')
    return
  }
  if (listening.value && recognition) {
    recognition.stop()
    return
  }
  recognition = new SR()
  recognition.lang = 'fr-FR'
  recognition.interimResults = false
  recognition.onresult = (ev: any) => {
    const text = ev.results?.[0]?.[0]?.transcript
    if (text) sendTurn(text)
  }
  recognition.onend = () => { listening.value = false }
  recognition.start()
  listening.value = true
}

async function hangUp(outcome: string) {
  stopTimer()
  await finalizeCall(outcome)
}

async function finalizeCall(outcome: string) {
  if (phase.value === 'analyzing' || phase.value === 'done') return
  phase.value = 'analyzing'
  stopTimer()
  const blob = await stopRecordingAsync(true)
  try {
    await uploadRecording(blob)
    const res: any = await $fetch(`/api/commercial/pitch-sims/${simId.value}/finalize`, {
      method: 'POST',
      body: {
        outcome,
        durationSec: 8 * 60 - remainingSec.value,
        transcript: transcript.value,
      },
    })
    const data = res.data ?? res
    if (typeof data.coachFeedback === 'string') {
      try { data.coachFeedback = JSON.parse(data.coachFeedback) } catch { /* keep */ }
    }
    if (localAudioUrl.value && !data.audioUrl) data.audioUrl = localAudioUrl.value
    sim.value = data
    phase.value = 'done'
    feedbackDone.value = !!data.hasFeedback || !!data.feedbackSkipped
    await loadSkipQuota()
  } catch {
    phase.value = 'done'
    sim.value = { outcome, aiScore: null, coachFeedback: null, audioUrl: localAudioUrl.value }
  }
}

async function saveUserScore() {
  if (userScore.value == null || !simId.value) return
  await $fetch(`/api/commercial/pitch-sims/${simId.value}/rating`, {
    method: 'PATCH',
    body: { score: userScore.value },
  })
}

async function submitFeedback(skip: boolean) {
  if (!simId.value) return
  await $fetch(`/api/commercial/pitch-sims/${simId.value}/feedback`, {
    method: 'POST',
    body: skip
      ? { skip: true }
      : {
          vetRealism: fb.vetRealism,
          coachUsefulness: fb.coachUsefulness,
          difficultyFelt: fb.difficultyFelt,
          comment: fb.comment,
          flags: [],
        },
  })
  feedbackDone.value = true
}

function resetCall() {
  phase.value = 'idle'
  simId.value = ''
  sim.value = null
  transcript.value = []
  feedbackDone.value = false
  remainingSec.value = 8 * 60
  error.value = ''
  if (localAudioUrl.value) {
    URL.revokeObjectURL(localAudioUrl.value)
    localAudioUrl.value = ''
  }
}

onMounted(async () => {
  if (typeof window !== 'undefined' && window.speechSynthesis) {
    window.speechSynthesis.getVoices()
    window.speechSynthesis.onvoiceschanged = () => window.speechSynthesis.getVoices()
  }
  await loadScripts()
  await loadSkipQuota()
})

onBeforeUnmount(() => {
  stopTimer()
  void stopRecordingAsync(false)
})
</script>

<style scoped>
.pf-training-tabs { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.pf-training-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  align-items: start;
}
.pf-steps { padding-left: 1.2rem; display: grid; gap: 0.75rem; }
.pf-steps ul { margin: 0.25rem 0 0; padding-left: 1rem; }
.pf-example { margin-top: 1rem; }
.pf-dialogue-line { margin: 0.35rem 0; font-size: 0.9rem; }
.pf-difficulty {
  display: grid;
  gap: 0.5rem;
  margin: 0.75rem 0;
}
.pf-diff-card {
  text-align: left;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.65rem 0.75rem;
  background: var(--pf-vet-surface);
  cursor: pointer;
  display: grid;
  gap: 0.2rem;
  font: inherit;
  color: inherit;
}
.pf-diff-card--active { border-color: var(--pf-vet-accent); box-shadow: var(--pf-vet-shadow-sm); }
.pf-diff-card span { color: var(--pf-vet-muted, #64748b); font-size: 0.85rem; }
.pro-select, .pro-input, .pro-label {
  display: block;
  width: 100%;
  margin-top: 0.35rem;
}
.pro-label { font-size: 0.85rem; font-weight: 600; margin-top: 0.75rem; }
.pro-select, .pro-input {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.5rem 0.65rem;
  background: #fff;
}
.pro-mt-md { margin-top: 1rem; }
.pro-mb-lg { margin-bottom: 1.25rem; }
.pf-ringing { text-align: center; padding: 2rem 1rem; display: grid; gap: 0.75rem; justify-items: center; }
.pf-call-bar { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; margin-bottom: 0.75rem; }
.pf-countdown { font-variant-numeric: tabular-nums; font-size: 1.25rem; }
.pf-transcript {
  max-height: 280px;
  overflow: auto;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.75rem;
  display: grid;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  background: var(--pf-vet-bg, #f8fafc);
}
.pf-line--vet strong { color: var(--pf-vet-primary); }
.pf-line--commercial strong { color: var(--pf-vet-accent); }
.pf-mic-row { display: grid; gap: 0.5rem; }
.pf-text-turn { display: flex; gap: 0.5rem; }
.pf-error { color: var(--pf-vet-alert); margin-top: 0.5rem; }
.pf-script-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.pf-fb-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.pf-audio { width: 100%; margin-top: 0.35rem; }
.pf-audio-mini { max-width: 140px; height: 28px; vertical-align: middle; }
@media (max-width: 900px) {
  .pf-training-grid { grid-template-columns: 1fr; }
}
</style>
