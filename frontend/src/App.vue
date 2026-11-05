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
      throw new Error(data.error || 'Error processing the demo')
    }

    matchData.value = data
  } catch (err) {
    errorMsg.value = err.message
  } finally {
    isLoading.value = false
  }
}

const columns = [
  { key: 'kills', label: 'K', color: 'text-emerald-400' },
  { key: 'deaths', label: 'D', color: 'text-rose-400' },
  { key: 'assists', label: 'A', color: 'text-zinc-300' },
  { key: 'headshot_percentage', label: 'HS%', color: 'text-amber-300', suffix: '%' },
  { key: 'total_damage', label: 'DMG', color: 'text-white font-semibold' },
  { key: 'utility_damage', label: 'Utility DMG', color: 'text-zinc-300' },
  { key: 'enemies_flashed', label: 'Flashed', color: 'text-sky-300' },
  // Opening duels: first kill of each round. "duel" cells render kills–deaths, sorted by kills.
  {
    key: 'opening_kills',
    label: 'Entry K–D',
    title: 'Opening duels: first kills – first deaths of a round',
    type: 'duel',
    divider: true
  },
  {
    key: 'opening_success',
    label: 'Entry %',
    title: 'Share of opening duels won',
    type: 'entryRate'
  }
]

const openingDuels = (player) => (player.opening_kills ?? 0) + (player.opening_deaths ?? 0)

// Green when the player wins most of his opening duels, red when he loses most of them
const entryRateClass = (player) => {
  const rate = player.opening_success ?? 0
  if (rate > 50) return 'text-emerald-400'
  if (rate < 50) return 'text-rose-400'
  return 'text-zinc-300'
}

// Visual identity of each side; the backend only sends "CT" / "T"
const sides = {
  CT: {
    fallbackName: 'Counter-Terrorists',
    label: 'CT',
    text: 'text-sky-400',
    badge: 'bg-sky-400/10 text-sky-300 ring-sky-400/30',
    bar: 'bg-sky-400'
  },
  T: {
    fallbackName: 'Terrorists',
    label: 'T',
    text: 'text-amber-400',
    badge: 'bg-amber-400/10 text-amber-300 ring-amber-400/30',
    bar: 'bg-amber-400'
  }
}

const sortPlayers = (players) => {
  const key = sortKey.value
  return [...players].sort((a, b) => (b[key] ?? 0) - (a[key] ?? 0))
}

const teamName = (side) => matchData.value?.[side.toLowerCase()]?.name || sides[side].fallbackName
const teamScore = (side) => matchData.value?.[side.toLowerCase()]?.score ?? 0

// One entry per side, winner first (like the in-game scoreboard). Ties keep CT on top.
const teams = computed(() => {
  if (!matchData.value || !matchData.value.players) return []
  const all = Object.values(matchData.value.players)

  const result = ['CT', 'T'].map((side) => ({
    side,
    name: teamName(side),
    score: teamScore(side),
    players: sortPlayers(all.filter((p) => p.team === side))
  }))

  return result.sort((a, b) => b.score - a.score)
})

const winnerSide = computed(() => {
  const ct = teamScore('CT')
  const t = teamScore('T')
  if (ct === t) return null
  return ct > t ? 'CT' : 'T'
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
          class="grid grid-cols-[1fr_auto_1fr] items-center gap-4 rounded-xl border border-zinc-800 bg-zinc-900 px-6 py-5"
        >
          <div class="min-w-0">
            <p class="truncate text-xs font-medium" :class="sides.CT.text">{{ teamName('CT') }}</p>
            <p
              class="text-4xl font-bold tabular-nums"
              :class="winnerSide && winnerSide !== 'CT' ? 'text-zinc-500' : 'text-white'"
            >
              {{ teamScore('CT') }}
            </p>
          </div>
          <span class="rounded-full bg-zinc-800 px-4 py-1 text-sm font-medium text-zinc-200">
            {{ matchData.map_name || 'Unknown map' }}
          </span>
          <div class="min-w-0 text-right">
            <p class="truncate text-xs font-medium" :class="sides.T.text">{{ teamName('T') }}</p>
            <p
              class="text-4xl font-bold tabular-nums"
              :class="winnerSide && winnerSide !== 'T' ? 'text-zinc-500' : 'text-white'"
            >
              {{ teamScore('T') }}
            </p>
          </div>
        </div>

        <p class="text-xs text-zinc-500">Click a column to sort both teams.</p>

        <!-- One scoreboard per team -->
        <div
          v-for="team in teams"
          :key="team.side"
          class="overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900"
        >
          <div class="relative flex items-center justify-between gap-3 border-b border-zinc-800 px-6 py-3">
            <span class="absolute inset-y-0 left-0 w-1" :class="sides[team.side].bar"></span>
            <div class="flex min-w-0 items-center gap-3">
              <span
                class="rounded-md px-2 py-0.5 text-xs font-bold ring-1 ring-inset"
                :class="sides[team.side].badge"
              >
                {{ sides[team.side].label }}
              </span>
              <h2 class="truncate font-semibold">{{ team.name }}</h2>
              <span
                v-if="winnerSide === team.side"
                class="rounded-full bg-[#FF5533]/15 px-2 py-0.5 text-xs font-medium text-[#FF5533]"
              >
                Winner
              </span>
            </div>
            <span class="text-2xl font-bold tabular-nums text-zinc-200">{{ team.score }}</span>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full min-w-[880px] table-fixed text-left text-sm">
              <thead>
                <tr class="bg-zinc-950/50 text-xs text-zinc-400">
                  <th class="w-52 px-6 py-3 font-medium">Player</th>
                  <th
                    v-for="col in columns"
                    :key="col.key"
                    class="cursor-pointer px-4 py-3 text-center font-medium whitespace-nowrap select-none hover:text-zinc-100"
                    :class="[
                      sortKey === col.key ? 'text-[#FF5533]' : '',
                      col.divider ? 'border-l border-zinc-800' : ''
                    ]"
                    :title="col.title"
                    @click="sortKey = col.key"
                  >
                    {{ col.label }}<span v-if="sortKey === col.key"> ↓</span>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-zinc-800">
                <tr v-if="team.players.length === 0">
                  <td :colspan="columns.length + 1" class="px-6 py-4 text-center text-zinc-500">
                    No players found for this side
                  </td>
                </tr>
                <tr
                  v-for="player in team.players"
                  :key="player.steam_id"
                  class="transition-colors hover:bg-zinc-800/60"
                >
                  <td class="truncate px-6 py-3 font-medium text-white" :title="player.name">{{ player.name }}</td>
                  <td
                    v-for="col in columns"
                    :key="col.key"
                    class="px-4 py-3 text-center tabular-nums whitespace-nowrap"
                    :class="[col.color, col.divider ? 'border-l border-zinc-800' : '']"
                  >
                    <span v-if="col.type === 'duel' && openingDuels(player) === 0" class="text-zinc-600">0 – 0</span>
                    <template v-else-if="col.type === 'duel'">
                      <span class="text-emerald-400">{{ player.opening_kills ?? 0 }}</span>
                      <span class="text-zinc-600"> – </span>
                      <span class="text-rose-400">{{ player.opening_deaths ?? 0 }}</span>
                    </template>
                    <template v-else-if="col.type === 'entryRate'">
                      <span v-if="openingDuels(player) === 0" class="text-zinc-600" title="No opening duels">—</span>
                      <span v-else :class="entryRateClass(player)">{{ player.opening_success }}%</span>
                    </template>
                    <template v-else>{{ player[col.key] ?? 0 }}{{ col.suffix || '' }}</template>
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