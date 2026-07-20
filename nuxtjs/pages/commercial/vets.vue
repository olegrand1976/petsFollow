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
        <ProInput v-model="form.prospectId" test-id="encode-vet-prospect" :label="$t('commercial.vets.prospectId')" />
        <p v-if="formError" class="pro-error">{{ formError }}</p>
        <ProButton type="submit" test-id="encode-vet-submit" :disabled="saving">{{ $t('commercial.vets.save') }}</ProButton>
      </form>
    </ProCard>

    <ProCard class="pro-mb-lg" data-testid="commercial-client-form">
      <h3 class="pro-mb-md">{{ $t('commercial.clients.create') }}</h3>
      <p class="pro-hint pro-mb-md">{{ $t('commercial.clients.hint') }}</p>
      <form class="pro-form" @submit.prevent="submitClient">
        <div class="pro-field">
          <label class="pro-label" for="client-vet">{{ $t('commercial.clients.vet') }}</label>
          <select id="client-vet" v-model="clientForm.vetUserId" class="pro-select" required data-testid="create-client-vet">
            <option value="">{{ $t('commercial.clients.vetPlaceholder') }}</option>
            <option v-for="v in vets" :key="v.userId" :value="v.userId">
              {{ v.fullName }} — {{ v.practiceName }}
            </option>
          </select>
        </div>
        <ProInput v-model="clientForm.fullName" test-id="create-client-name" :label="$t('commercial.clients.fullName')" required />
        <ProInput v-model="clientForm.email" test-id="create-client-email" type="email" :label="$t('commercial.clients.email')" required />
        <ProInput v-model="clientForm.password" test-id="create-client-password" type="password" :label="$t('commercial.clients.password')" required />
        <p v-if="clientMsg" class="pro-hint">{{ clientMsg }}</p>
        <p v-if="clientError" class="pro-error">{{ clientError }}</p>
        <ProButton type="submit" test-id="create-client-submit" :disabled="clientSaving">{{ $t('commercial.clients.submit') }}</ProButton>
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
  prospectId: '',
})
const clientForm = reactive({ vetUserId: '', fullName: '', email: '', password: '' })
const clientSaving = ref(false)
const clientMsg = ref('')
const clientError = ref('')

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
    Object.assign(form, { fullName: '', practiceName: '', email: '', password: '', phone: '', city: '', postalCode: '', addressLine1: '', contactEmail: '', prospectId: '' })
    await load()
  } catch {
    formError.value = t('commercial.vets.encodeFailed')
  } finally {
    saving.value = false
  }
}

async function submitClient() {
  clientSaving.value = true
  clientMsg.value = ''
  clientError.value = ''
  try {
    await $fetch('/api/commercial/clients', { method: 'POST', body: { ...clientForm } })
    clientMsg.value = t('commercial.clients.success')
    Object.assign(clientForm, { vetUserId: '', fullName: '', email: '', password: '' })
    await load()
  } catch {
    clientError.value = t('commercial.clients.error')
  } finally {
    clientSaving.value = false
  }
}

onMounted(load)
</script>
