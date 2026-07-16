<template>
  <div data-testid="commercial-vets-page">
    <ProPageHeader
      :title="$t('commercial.vets.title')"
      :subtitle="$t('commercial.vets.subtitle')"
    />

    <ProCard class="pro-mb-lg" data-testid="commercial-vet-form">
      <h3 class="pro-mb-md">{{ $t('commercial.vets.encode') }}</h3>
      <form class="pro-form" @submit.prevent="submitVet">
        <ProInput v-model="form.fullName" test-id="encode-vet-name" :label="$t('commercial.vets.fullName')" required />
        <ProInput v-model="form.practiceName" test-id="encode-vet-practice" :label="$t('commercial.vets.practiceName')" required />
        <ProInput v-model="form.email" test-id="encode-vet-email" type="email" :label="$t('commercial.vets.email')" required />
        <ProInput v-model="form.password" test-id="encode-vet-password" type="password" :label="$t('commercial.vets.password')" required />
        <ProInput v-model="form.phone" test-id="encode-vet-phone" :label="$t('commercial.vets.phone')" />
        <ProInput v-model="form.city" test-id="encode-vet-city" :label="$t('commercial.vets.city')" />
        <ProInput v-model="form.postalCode" test-id="encode-vet-postal" :label="$t('commercial.vets.postalCode')" />
        <ProInput v-model="form.addressLine1" test-id="encode-vet-address" :label="$t('commercial.vets.address')" />
        <p v-if="formError" class="pro-error">{{ formError }}</p>
        <ProButton type="submit" test-id="encode-vet-submit" :disabled="saving">{{ $t('commercial.vets.save') }}</ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProTable :empty="!vets.length" :empty-title="$t('commercial.vets.empty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.vets.fullName') }}</th>
            <th>{{ $t('commercial.vets.email') }}</th>
            <th>{{ $t('commercial.vets.practiceName') }}</th>
            <th>{{ $t('commercial.vets.clients') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in vets" :key="v.userId" :data-testid="`commercial-vet-row-${v.userId}`">
            <td>{{ v.fullName }}</td>
            <td>{{ v.email }}</td>
            <td>{{ v.practiceName }}</td>
            <td>{{ v.clientCount }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()
const vets = ref<any[]>([])
const saving = ref(false)
const formError = ref('')
const form = reactive({
  fullName: '',
  practiceName: '',
  email: '',
  password: '',
  phone: '',
  city: '',
  postalCode: '',
  addressLine1: '',
  contactEmail: '',
})

async function load() {
  const res: any = await $fetch('/api/commercial/vets')
  vets.value = res.data ?? res ?? []
}

async function submitVet() {
  saving.value = true
  formError.value = ''
  try {
    await $fetch('/api/commercial/vets', {
      method: 'POST',
      body: { ...form, contactEmail: form.contactEmail || form.email },
    })
    Object.assign(form, { fullName: '', practiceName: '', email: '', password: '', phone: '', city: '', postalCode: '', addressLine1: '', contactEmail: '' })
    await load()
  } catch {
    formError.value = t('commercial.vets.encodeFailed')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
