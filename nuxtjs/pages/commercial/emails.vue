<template>
  <div data-testid="commercial-emails-page">
    <ProPageHeader
      :title="$t('commercial.mail.emailsTitle')"
      :subtitle="$t('commercial.mail.emailsSubtitle')"
    />
    <p v-if="error" class="pro-field-error pro-mb-md" role="alert">{{ error }}</p>

    <ProCard>
      <ProTable :empty="!items.length" :empty-title="$t('commercial.mail.emptySends')">
        <thead>
          <tr>
            <th>{{ $t('commercial.mail.date') }}</th>
            <th>{{ $t('commercial.mail.practice') }}</th>
            <th>{{ $t('commercial.mail.subject') }}</th>
            <th>{{ $t('commercial.mail.to') }}</th>
            <th>{{ $t('commercial.mail.status') }}</th>
            <th>{{ $t('commercial.mail.opened') }}</th>
            <th>{{ $t('commercial.mail.clicks') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in items" :key="s.id" :data-testid="`email-send-row-${s.id}`">
            <td>{{ formatDate(s.sentAt || s.createdAt) }}</td>
            <td>{{ s.practiceName || '—' }}</td>
            <td>{{ s.subject }}</td>
            <td>{{ s.toEmail }}</td>
            <td>
              <ProBadge :variant="s.status === 'sent' ? 'success' : 'danger'">
                {{ $t(`commercial.mail.sendStatus.${s.status}`) }}
              </ProBadge>
            </td>
            <td>
              <template v-if="s.openCount > 0">{{ s.openCount }} · {{ formatDate(s.openedAt) }}</template>
              <span v-else class="pro-hint">{{ $t('commercial.mail.neverOpened') }}</span>
            </td>
            <td>{{ s.clickTotal ?? 0 }}</td>
            <td>
              <ProButton variant="ghost" :test-id="`email-send-detail-${s.id}`" @click="openDetail(s.id)">
                {{ $t('commercial.mail.detail') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard v-if="detail" class="pro-mt-lg" data-testid="email-send-detail">
      <strong>{{ detail.subject }}</strong>
      <p class="pro-hint">{{ detail.toEmail }} · {{ detail.practiceName }}</p>
      <p>
        {{ $t('commercial.mail.opened') }}:
        <template v-if="detail.openCount > 0">{{ detail.openCount }} · {{ formatDate(detail.openedAt) }}</template>
        <span v-else>{{ $t('commercial.mail.neverOpened') }}</span>
      </p>
      <div v-if="detail.clicks?.length" class="pro-mt-md">
        <strong>{{ $t('commercial.mail.clickList') }}</strong>
        <ul>
          <li v-for="c in detail.clicks" :key="c.id">
            {{ c.targetUrl }} — {{ c.clickCount }}×
            <span v-if="c.clickedAt">({{ formatDate(c.clickedAt) }})</span>
          </li>
        </ul>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'commercial-only' })

const { mapError } = useApiError()
const items = ref<any[]>([])
const detail = ref<any | null>(null)
const error = ref('')

function formatDate(iso?: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString()
}

async function load() {
  error.value = ''
  try {
    const res: any = await $fetch('/api/commercial/emails')
    const data = res.data ?? res
    items.value = data.items ?? []
  } catch (e: any) {
    error.value = mapError(e)
  }
}

async function openDetail(id: string) {
  error.value = ''
  try {
    const res: any = await $fetch(`/api/commercial/emails/${id}`)
    detail.value = res.data ?? res
  } catch (e: any) {
    error.value = mapError(e)
  }
}

onMounted(load)
</script>
