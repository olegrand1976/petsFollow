<template>
  <div>
    <ProPageHeader
      title="Paramètres"
      subtitle="Disponibilité et messages automatiques."
    />
    <ProCard title="Disponibilité">
      <div class="pro-toggle" role="group" aria-label="Statut de disponibilité">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': status === 'available' }"
          @click="status = 'available'"
        >
          Disponible
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': status === 'unavailable' }"
          @click="status = 'unavailable'"
        >
          Indisponible
        </button>
      </div>
      <div class="pro-field pro-field-spaced">
        <label class="pro-label" for="auto-reply">Message auto-réponse</label>
        <textarea
          id="auto-reply"
          v-model="autoReply"
          class="pro-textarea"
          rows="4"
          placeholder="Message envoyé automatiquement lorsque vous êtes indisponible."
        />
      </div>
      <p v-if="saved" class="text-muted" role="status">Paramètres enregistrés.</p>
      <ProButton class="pro-save-btn" :loading="saving" @click="save">
        Enregistrer
      </ProButton>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const status = ref('available')
const autoReply = ref('Je suis indisponible, je reviens vers vous rapidement.')
const saving = ref(false)
const saved = ref(false)

onMounted(async () => {
  const res: any = await $fetch('/api/vet/availability')
  const data = res.data ?? res
  status.value = data.status ?? status.value
  autoReply.value = data.autoReply || autoReply.value
})

async function save() {
  saving.value = true
  saved.value = false
  try {
    await $fetch('/api/vet/availability', {
      method: 'PUT',
      body: { status: status.value, autoReply: autoReply.value },
    })
    saved.value = true
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.pro-field-spaced {
  margin-top: 1.25rem;
}

.pro-save-btn {
  margin-top: 1rem;
}
</style>
