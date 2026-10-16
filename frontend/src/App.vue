<script setup>
import { ref, computed } from 'vue'

// Reactive state for file upload and parsing status
const selectedFile = ref(null)
const isLoading = ref(false)
const matchData = ref(null)
const errorMsg = ref('')
const isDragging = ref(false)

// Sorting state and animation trigger for stat bars
const sortKey = ref('total_damage')
const sortDesc = ref(true)
const barsReady = ref(false)

// Validate and store the selected .dem file
const setFile = (file) => {
  if (!file) return
  if (!file.name.toLowerCase().endsWith('.dem')) {
    errorMsg.value = 'Only .dem files are supported. Pick a CS2 demo file.'
    return
  }
  selectedFile.value = file
  errorMsg.value = ''
}

// Handle standard file input selection
const onFileChange = (event) => setFile(event.target.files[0])

// Handle drag-and-drop file selection
const onDrop = (event) => {
  isDragging.value = false
  setFile(event.dataTransfer.files[0])
}

// Upload demo file to the Go backend API
const uploadDemo = async () => {
  if (!selectedFile.value) {
    errorMsg.value = 'Choose a .dem file first.'
    return
  }

  isLoading.value = true
  errorMsg.value = ''
  matchData.value = null
  barsReady.value = false

  const formData = new FormData()
  formData.append('demo', selectedFile.value)

  try {
    const response = await fetch('http://localhost:3000/api/upload', {
      method: 'POST',
      body: formData
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.eroare || 'The server could not process this demo.')
    }

    matchData.value = data
    // Trigger progress bar width transitions shortly after DOM render
    setTimeout(() => (barsReady.value = true), 60)
  } catch (err) {
    errorMsg.value = err.message
  } finally {
    isLoading.value = false
  }
}

// Sortable table columns configuration
const columns = [
  { key: 'kills', label: 'K / D / A' },
  { key: 'headshot_percentage', label: 'HS %' },
  { key: 'total_damage', label: 'Total Damage' },
  { key: 'utility_damage', label: 'Util. DMG' },
  { key: 'enemies_flashed', label: 'Flashed' }
]

// Change active sort column or toggle direction if clicking the same column
const toggleSort = (key) => {
  if (sortKey.value === key) {
    sortDesc.value = !sortDesc.value
  } else {
    sortKey.value = key
    sortDesc.value = true
  }
}

// Return players sorted by the currently selected metric
const sortedPlayers = computed(() => {
  if (!matchData.value || !matchData.value.players) return []
  const key = sortKey.value
  const dir = sortDesc.value ? 1 : -1

  return Object.values(matchData.value.players).sort(
    (a, b) => ((b[key] ?? 0) - (a[key] ?? 0)) * dir
  )
})

// Highest damage in the lobby used to scale player damage bars
const maxDamage = computed(() =>
  Math.max(1, ...sortedPlayers.value.map((p) => p.total_damage ?? 0))
)

// Calculate CT round win share for the dominance bar
const ctShare = computed(() => {
  const ct = matchData.value?.ct_score ?? 0
  const t = matchData.value?.t_score ?? 0
  return ct + t === 0 ? 50 : (ct / (ct + t)) * 100
})

// Format selected file size in megabytes
const fileSize = computed(() =>
  selectedFile.value ? (selectedFile.value.size / 1048576).toFixed(1) + ' MB' : ''
)

// Calculate K/D ratio formatted to 2 decimal places
const formatKD = (kills = 0, deaths = 0) => {
  if (deaths === 0) return kills.toFixed(2)
  return (kills / deaths).toFixed(2)
}
</script>

<template>
  <div class="min-h-screen bg-[#090b10] font-sans text-slate-100 selection:bg-rose-500/30 selection:text-rose-200">
    <!-- Top ambient glow background effect -->
    <div class="pointer-events-none fixed inset-x-0 top-0 h-96 bg-gradient-to-b from-rose-500/5 via-[#FF775C]/5 to-transparent"></div>

    <div class="relative mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <!-- Header & Upload Command Bar -->
      <header class="mb-8 rounded-2xl border border-slate-800/80 bg-[#11151f]/90 p-6 shadow-2xl backdrop-blur-md">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
          <!-- Brand Identity -->
          <div>
            <div class="flex items-center gap-3">
              <span class="inline-flex h-3 w-3 rounded-full bg-rose-400 shadow-[0_0_12px_#F6658C]"></span>
              <span class="text-xs font-bold tracking-[0.2em] text-rose-400 uppercase">CS2 Telemetry Engine</span>
            </div>
            <h1 class="mt-1 text-3xl font-black tracking-tight text-white sm:text-4xl">
              CS2 <span class="text-[#F6658C]">DEMO</span> ANALYZER
            </h1>
            <p class="mt-1 text-sm text-slate-400">
              Upload a CS2 replay file (<code class="text-slate-300">.dem</code>) for instant combat & utility breakdown.
            </p>
          </div>

          <!-- Upload Controls -->
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
            <label
              class="group relative flex min-w-[260px] cursor-pointer items-center gap-3.5 rounded-xl border-2 px-4 py-3 transition-all"
              :class="[
                isDragging
                  ? 'border-rose-400 bg-rose-500/10'
                  : selectedFile
                    ? 'border-emerald-500/50 bg-emerald-500/5 hover:border-emerald-400'
                    : 'border-dashed border-slate-700 bg-[#0c0f17] hover:border-slate-500 hover:bg-slate-800/40'
              ]"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="onDrop"
            >
              <input type="file" accept=".dem" class="sr-only" @change="onFileChange" />

              <!-- Icon -->
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-lg font-bold"
                :class="selectedFile ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-800'"
              >
                {{ selectedFile ? '✓' : '↑' }}
              </div>

              <!-- File details -->
              <div class="min-w-0 flex-1">
                <template v-if="selectedFile">
                  <p class="truncate text-sm font-semibold text-white">{{ selectedFile.name }}</p>
                  <p class="text-xs text-emerald-400">{{ fileSize }} · Ready to parse</p>
                </template>
                <template v-else>
                  <p class="text-sm font-semibold text-slate-200 group-hover:text-white">Drop .dem file here</p>
                  <p class="text-xs text-slate-400">or click to browse local files</p>
                </template>
              </div>
            </label>

            <button
              class="cursor-pointer rounded-xl bg-gradient-to-r from-rose-400 to-rose-300 px-6 py-3.5 text-sm font-bold tracking-wide text-slate-950 uppercase shadow-lg shadow-rose-500/20 transition-all hover:shadow-rose-400/30 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-rose-400 disabled:cursor-not-allowed disabled:from-slate-800 disabled:to-slate-800 disabled:text-slate-500 disabled:shadow-none"
              :disabled="isLoading"
              @click="uploadDemo"
            >
              {{ isLoading ? 'Parsing Ticks…' : 'Analyze Match' }}
            </button>
          </div>
        </div>

        <!-- Error Alert Banner -->
        <div
          v-if="errorMsg"
          class="mt-4 flex items-center gap-3 rounded-xl border border-rose-500/40 bg-rose-500/10 px-4 py-3 text-sm text-rose-200"
          role="alert"
        >
          <span class="font-bold text-rose-400">ERROR:</span>
          <span>{{ errorMsg }}</span>
        </div>
      </header>

      <!-- Main Content Stage -->
      <main>
        <!-- Loading State -->
        <div
          v-if="isLoading"
          class="flex flex-col items-center justify-center rounded-2xl border border-slate-800/80 bg-[#11151f]/60 py-24 text-center"
          aria-live="polite"
        >
          <div class="h-12 w-12 animate-spin rounded-full border-4 border-rose-400 border-t-transparent"></div>
          <p class="mt-5 text-lg font-bold text-white">Extracting Match Telemetry…</p>
          <p class="mt-1 max-w-md text-sm text-slate-400">
            Scanning game events, calculating damage, and syncing final scoreboard data.
          </p>
        </div>

        <!-- Empty State -->
        <div
          v-else-if="!matchData"
          class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-800 bg-[#11151f]/30 py-24 text-center"
        >
          
          <p class="text-xl font-bold text-slate-200">No Match Telemetry Loaded</p>
          <p class="mt-1 max-w-sm text-sm text-slate-400">
            Select a Counter-Strike 2 demo file in the command bar above to generate the scoreboard.
          </p>
        </div>

        <!-- Loaded Match Dashboard -->
        <div v-else class="space-y-6">
          <!-- Match Scoreboard HUD Banner -->
          <section
            class="relative overflow-hidden rounded-2xl border border-slate-800 bg-[#11151f] p-6 shadow-xl sm:p-8"
            aria-label="Match score"
          >
            <div class="grid grid-cols-1 items-center gap-6 md:grid-cols-3">
              <!-- Counter-Terrorists Score -->
              <div class="flex items-center gap-5">
                <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl border border-blue-500/30 bg-blue-700/10 text-2xl font-black text-blue-600">
                  CT
                </div>
                <div>
                  <span class="text-xs font-bold tracking-widest text-blue-400 uppercase">Counter-Terrorists</span>
                  <p class="text-5xl font-black tracking-tight text-white tabular-nums sm:text-6xl">
                    {{ matchData.ct_score ?? 0 }}
                  </p>
                </div>
              </div>

              <!-- Map Info Pill -->
              <div class="flex flex-col items-center justify-center text-center">
                <span class="text-[11px] font-bold tracking-[0.2em] text-slate-400 uppercase">Active Map</span>
                <div class="mt-1.5 inline-flex items-center gap-2 rounded-full border border-slate-700 bg-[#090b10] px-5 py-2 text-lg font-extrabold text-white shadow-inner">
                  <span class="h-2 w-2 rounded-full bg-emerald-400"></span>
                  {{ matchData.map_name || 'Unknown Map' }}
                </div>
              </div>

              <!-- Terrorists Score -->
              <div class="flex items-center justify-end gap-5 text-right">
                <div>
                  <span class="text-xs font-bold tracking-widest text-amber-400 uppercase">Terrorists</span>
                  <p class="text-5xl font-black tracking-tight text-white tabular-nums sm:text-6xl">
                    {{ matchData.t_score ?? 0 }}
                  </p>
                </div>
                <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl border border-amber-500/30 bg-amber-500/10 text-2xl font-black text-amber-400">
                  T
                </div>
              </div>
            </div>

            <!-- Round Dominance Ratio Bar -->
            <div class="mt-6">
              <div class="mb-1.5 flex justify-between text-xs font-semibold text-slate-400">
                <span>CT Round Share ({{ Math.round(ctShare) }}%)</span>
                <span>T Round Share ({{ Math.round(100 - ctShare) }}%)</span>
              </div>
              <div class="flex h-2.5 w-full overflow-hidden rounded-full bg-amber-500/80 p-0.5" aria-hidden="true">
                <div
                  class="h-full rounded-full bg-blue-600 shadow-[0_0_10px_#2E2EFF] transition-[width] duration-700 ease-out motion-reduce:transition-none"
                  :style="{ width: ctShare + '%' }"
                ></div>
              </div>
            </div>
          </section>

          <!-- Player Telemetry Table -->
          <section class="overflow-hidden rounded-2xl border border-slate-800 bg-[#11151f] shadow-xl" aria-label="Players">
            <!-- Section Header & Quick Sort Pills for Mobile -->
            <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-800/80 bg-[#0d1018] px-6 py-4">
              <div>
                <h2 class="text-base font-bold text-white">Player Performance Breakdown</h2>
                <p class="text-xs text-slate-400">Click any column header to sort players by that metric.</p>
              </div>

              <!-- Mobile sort buttons -->
              <div class="flex flex-wrap gap-1.5 lg:hidden">
                <button
                  v-for="col in columns"
                  :key="col.key"
                  class="cursor-pointer rounded-lg px-2.5 py-1 text-xs font-semibold transition-colors"
                  :class="
                    sortKey === col.key
                      ? 'bg-rose-500/20 text-rose-300 border border-rose-500/40'
                      : 'bg-slate-800/70 text-slate-400 hover:text-slate-200'
                  "
                  @click="toggleSort(col.key)"
                >
                  {{ col.label }}
                  <span v-if="sortKey === col.key">{{ sortDesc ? '↓' : '↑' }}</span>
                </button>
              </div>
            </div>

            <!-- Desktop Column Headers -->
            <div
              class="hidden grid-cols-[minmax(180px,1.5fr)_140px_100px_minmax(180px,1.8fr)_100px_90px] items-center gap-4 border-b border-slate-800 bg-[#0b0e15] px-6 py-3.5 text-xs font-bold tracking-wider uppercase lg:grid"
            >
              <span class="text-slate-400">Player</span>
              <button
                v-for="col in columns"
                :key="col.key"
                class="flex cursor-pointer items-center gap-1.5 text-left font-bold tracking-wider uppercase transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-rose-400"
                :class="
                  sortKey === col.key
                    ? 'text-rose-400'
                    : 'text-slate-400 hover:text-slate-200'
                "
                @click="toggleSort(col.key)"
              >
                <span>{{ col.label }}</span>
                <span
                  v-if="sortKey === col.key"
                  class="inline-flex h-4 w-4 items-center justify-center rounded bg-rose-500/20 text-[11px] text-rose-300"
                >
                  {{ sortDesc ? '↓' : '↑' }}
                </span>
              </button>
            </div>

            <!-- Player Rows -->
            <ol class="m-0 divide-y divide-slate-800/60 list-none p-0">
              <li
                v-for="(p, i) in sortedPlayers"
                :key="p.steam_id"
                class="grid grid-cols-2 items-center gap-x-4 gap-y-3 px-6 py-4 transition-colors hover:bg-slate-800/30 sm:grid-cols-3 lg:grid-cols-[minmax(180px,1.5fr)_140px_100px_minmax(180px,1.8fr)_100px_90px]"
              >
                <!-- Rank & Player Name -->
                <div class="col-span-2 flex min-w-0 items-center gap-3.5 sm:col-span-3 lg:col-span-1">
                  <span
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg text-xs font-bold tabular-nums"
                    :class="[
                      i === 0
                        ? 'bg-amber-400/20 text-amber-300 border border-amber-400/40'
                        : i === 1
                          ? 'bg-slate-300/20 text-slate-200 border border-slate-300/30'
                          : i === 2
                            ? 'bg-amber-700/20 text-amber-500 border border-amber-700/40'
                            : 'bg-slate-800/80 text-slate-400'
                    ]"
                  >
                    #{{ i + 1 }}
                  </span>
                  <div class="min-w-0">
                    <p class="truncate text-base font-bold text-white">{{ p.name }}</p>
                  </div>
                </div>

                <!-- K/D/A -->
                <div class="tabular-nums">
                  <span class="mb-0.5 block text-[11px] font-semibold text-slate-400 uppercase lg:hidden">K / D / A</span>
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium">
                      <b class="font-bold text-emerald-400">{{ p.kills }}</b>
                      <i class="mx-1 not-italic text-slate-600">/</i>
                      <span class="text-rose-400">{{ p.deaths }}</span>
                      <i class="mx-1 not-italic text-slate-600">/</i>
                      <span class="text-slate-300">{{ p.assists }}</span>
                    </span>
                    <span
                      class="rounded px-1.5 py-0.5 text-[11px] font-bold"
                      :class="
                        (p.kills ?? 0) >= (p.deaths ?? 0)
                          ? 'bg-emerald-500/15 text-emerald-300'
                          : 'bg-rose-500/15 text-rose-300'
                      "
                    >
                      {{ formatKD(p.kills, p.deaths) }}
                    </span>
                  </div>
                </div>

                <!-- Headshot Percentage -->
                <div class="tabular-nums">
                  <span class="mb-0.5 block text-[11px] font-semibold text-slate-400 uppercase lg:hidden">HS %</span>
                  <span class="inline-flex items-center gap-1 text-sm font-semibold text-amber-300">
                    {{ p.headshot_percentage ?? 0 }}%
                  </span>
                </div>

                <!-- Total Damage -->
                <div class="col-span-2 flex items-center gap-3 sm:col-span-3 lg:col-span-1">
                  <div class="h-2 flex-1 overflow-hidden rounded-full bg-slate-800">
                    <div
                      class="h-full rounded-full bg-gradient-to-r from-rose-400 to-blue-500 transition-[width] duration-700 ease-out motion-reduce:transition-none"
                      :style="{
                        width: barsReady ? ((p.total_damage ?? 0) / maxDamage) * 100 + '%' : '0%'
                      }"
                    ></div>
                  </div>
                  <span class="w-12 text-right font-mono text-sm font-bold text-white tabular-nums">
                    {{ p.total_damage ?? 0 }}
                  </span>
                </div>

                <!-- Utility Damage -->
                <div class="tabular-nums">
                  <span class="mb-0.5 block text-[11px] font-semibold text-slate-400 uppercase lg:hidden">Utility DMG</span>
                  <span class="text-sm font-semibold text-purple-300">
                    {{ p.utility_damage ?? 0 }}
                  </span>
                </div>

                <!-- Enemies Flashed -->
                <div class="tabular-nums">
                  <span class="mb-0.5 block text-[11px] font-semibold text-slate-400 uppercase lg:hidden">Flashed</span>
                  <span class="inline-flex items-center gap-1.5 text-sm font-semibold text-sky-300">
                    {{ p.enemies_flashed ?? 0 }}
                  </span>
                </div>
              </li>
            </ol>
          </section>
        </div>
      </main>
    </div>
  </div>
</template>