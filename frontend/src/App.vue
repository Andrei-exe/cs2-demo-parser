<script setup>
import { ref, computed } from 'vue'

const selectedFile = ref(null)
const isLoading = ref(false)
const matchData = ref(null)
const errorMsg = ref('')
const sortKey = ref('total_damage')

const onFileChange = (event) => {
  selectedFile.value = event.target.files[0]
  errorMsg.value = ''
}

const uploadDemo = async () => {
  if (!selectedFile.value) {
    errorMsg.value = 'Please select a .dem file first'
    return
  }

  isLoading.value = true
  errorMsg.value = ''
  matchData.value = null

  const formData = new FormData()
  formData.append('demo', selectedFile.value)

  try {
    const response = await fetch('http://localhost:3000/api/upload', {
      method: 'POST',
      body: formData
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.eroare || 'Error processing the demo')
    }

    matchData.value = data
  } catch (err) {
    errorMsg.value = err.message
  } finally {
    isLoading.value = false
  }
}

const columns = [
  { key: 'kills', label: 'Kills', color: 'text-emerald-400' },
  { key: 'deaths', label: 'Deaths', color: 'text-rose-400' },
  { key: 'assists', label: 'Assists', color: 'text-zinc-300' },
  { key: 'headshot_percentage', label: 'HS%', color: 'text-amber-300', suffix: '%' },
  { key: 'total_damage', label: 'DMG', color: 'text-white font-semibold' },
  { key: 'utility_damage', label: 'Utility DMG', color: 'text-zinc-300' },
  { key: 'enemies_flashed', label: 'Enemies flashed', color: 'text-sky-300' }
]

const sortedPlayers = computed(() => {
  if (!matchData.value || !matchData.value.players) return []
  const key = sortKey.value
  return Object.values(matchData.value.players).sort((a, b) => (b[key] ?? 0) - (a[key] ?? 0))
})
</script>

<template>
  <div class="min-h-screen bg-zinc-950 px-4 py-10 text-zinc-100">
    <div class="mx-auto max-w-5xl">
      <!-- Header -->
      <header class="mb-8">
        <h1 class="text-3xl font-bold tracking-tight">CS2 Demo Analyzer</h1>
        <p class="mt-1 text-zinc-400">Upload a .dem file to see the match stats.</p>
      </header>

      <!-- Upload -->
      <div
        class="mb-6 flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-900 p-4 sm:flex-row sm:items-center"
      >
        <input
          type="file"
          accept=".dem"
          class="block w-full cursor-pointer text-sm text-zinc-400 file:mr-4 file:cursor-pointer file:rounded-lg file:border-0 file:bg-zinc-800 file:px-4 file:py-2 file:text-sm file:font-medium file:text-zinc-100 hover:file:bg-zinc-700"
          @change="onFileChange"
        />
        <button
          class="cursor-pointer rounded-lg bg-[#FF5533] text-zinc-950 px-5 py-2 text-sm font-semibold whitespace-nowrap transition hover:bg-[#ff6e52] disabled:cursor-wait disabled:bg-zinc-700 disabled:text-zinc-400"
          :disabled="isLoading"
          @click="uploadDemo"
        >
          {{ isLoading ? 'Analyzing...' : 'Analyze demo' }}
        </button>
      </div>

      <!-- Error -->
      <div
        v-if="errorMsg"
        class="mb-6 rounded-lg border border-rose-900 bg-rose-950/60 px-4 py-3 text-sm text-rose-200"
        role="alert"
      >
        {{ errorMsg }}
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="flex items-center gap-3 py-10 text-zinc-400">
        <div
          class="h-6 w-6 animate-spin rounded-full border-2 border-[#FF5533] border-t-transparent"
        ></div>
        <span>Analyzing match, this can take a while...</span>
      </div>

      <!-- Results -->
      <div v-if="matchData" class="space-y-6">
        <!-- Score -->
        <div
          class="flex items-center justify-between rounded-xl border border-zinc-800 bg-zinc-900 px-6 py-5"
        >
          <div>
            <p class="text-xs font-medium text-sky-400">Counter-Terrorists</p>
            <p class="text-4xl font-bold tabular-nums">{{ matchData.ct_score ?? 0 }}</p>
          </div>
          <span class="rounded-full bg-zinc-800 px-4 py-1 text-sm font-medium text-zinc-200">
            {{ matchData.map_name || 'Unknown map' }}
          </span>
          <div class="text-right">
            <p class="text-xs font-medium text-amber-400">Terrorists</p>
            <p class="text-4xl font-bold tabular-nums">{{ matchData.t_score ?? 0 }}</p>
          </div>
        </div>

        <!-- Scoreboard -->
        <div class="overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900">
          <div class="flex items-center justify-between border-b border-zinc-800 px-6 py-3">
            <h2 class="font-semibold">Scoreboard</h2>
            <span class="text-xs text-zinc-500">Click a column to sort</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead>
                <tr class="bg-zinc-950/50 text-xs text-zinc-400">
                  <th class="px-6 py-3 font-medium">Player</th>
                  <th
                    v-for="col in columns"
                    :key="col.key"
                    class="cursor-pointer px-4 py-3 text-center font-medium select-none hover:text-zinc-100"
                    :class="sortKey === col.key ? 'text-[#FF5533]' : ''"
                    @click="sortKey = col.key"
                  >
                    {{ col.label }}<span v-if="sortKey === col.key"> ↓</span>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-zinc-800">
                <tr
                  v-for="player in sortedPlayers"
                  :key="player.steam_id"
                  class="transition-colors hover:bg-zinc-800/60"
                >
                  <td class="px-6 py-3 font-medium text-white">{{ player.name }}</td>
                  <td
                    v-for="col in columns"
                    :key="col.key"
                    class="px-4 py-3 text-center tabular-nums"
                    :class="col.color"
                  >
                    {{ player[col.key] ?? 0 }}{{ col.suffix || '' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>