<template>
  <div data-testid="commercial-email-templates-page">
    <ProPageHeader
      :title="$t('commercial.mail.templatesTitle')"
      :subtitle="$t('commercial.mail.templatesSubtitle')"
    />
    <p v-if="error" class="pro-field-error pro-mb-md" role="alert">{{ error }}</p>
    <p v-if="savedOk" class="pro-hint pro-mb-md" role="status" data-testid="email-tpl-saved">{{ $t('commercial.mail.saved') }}</p>
    <p class="pro-hint pro-mb-md">{{ $t('commercial.mail.varsHint') }}</p>

    <div class="pf-mail-layout">
      <ProCard>
        <ProTable
          :empty="!templates.length"
          :empty-title="$t('commercial.mail.emptyTemplates')"
        >
          <thead>
            <tr>
              <th>{{ $t('commercial.mail.template') }}</th>
              <th>{{ $t('commercial.mail.category') }}</th>
              <th>{{ $t('commercial.mail.status') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="tpl in templates"
              :key="tpl.id"
              :data-testid="`email-tpl-row-${tpl.slug}`"
            >
              <td>{{ tpl.name }}</td>
              <td>{{ $t(`commercial.mail.categories.${tpl.category}`) }}</td>
              <td>
                <ProBadge :variant="tpl.isActive ? 'success' : 'neutral'">
                  {{ tpl.isActive ? $t('commercial.mail.active') : $t('commercial.mail.inactive') }}
                </ProBadge>
              </td>
              <td>
                <ProButton variant="ghost" :test-id="`email-tpl-edit-${tpl.slug}`" @click="select(tpl)">
                  {{ $t('commercial.mail.edit') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <ProCard v-if="edit" data-testid="email-tpl-editor">
        <form class="pro-form" @submit.prevent="save">
          <ProInput v-model="edit.name" test-id="email-tpl-name" :label="$t('commercial.mail.template')" />
          <label class="pro-label">{{ $t('commercial.mail.category') }}</label>
          <select v-model="edit.category" class="pro-select" data-testid="email-tpl-category">
            <option v-for="c in categories" :key="c" :value="c">
              {{ $t(`commercial.mail.categories.${c}`) }}
            </option>
          </select>
          <ProInput v-model="edit.subject" test-id="email-tpl-subject" :label="$t('commercial.mail.subject')" />
          <label class="pro-label">{{ $t('commercial.mail.body') }}</label>
          <textarea
            v-model="edit.bodyHtml"
            class="pro-input pf-mail-body"
            rows="14"
            data-testid="email-tpl-body"
          />
          <label class="pro-checkbox">
            <input v-model="edit.isActive" type="checkbox" data-testid="email-tpl-active">
            {{ $t('commercial.mail.active') }}
          </label>
          <div class="pf-mail-actions">
            <ProButton type="submit" test-id="email-tpl-save" :loading="saving">
              {{ $t('common.save') }}
            </ProButton>
            <ProButton variant="secondary" test-id="email-tpl-preview" :loading="previewing" @click="preview">
              {{ $t('commercial.mail.preview') }}
            </ProButton>
          </div>
        </form>
        <iframe
          v-if="previewHtml"
          class="pf-mail-preview"
          :title="$t('commercial.mail.preview')"
          data-testid="email-tpl-preview-frame"
          sandbox=""
          :srcdoc="previewHtml"
        />
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'commercial-only' })

const { mapError } = useApiError()
const categories = ['intro', 'rdv', 'nurture', 'post', 'reactivation'] as const
const templates = ref<any[]>([])
const edit = ref<any | null>(null)
const error = ref('')
const saving = ref(false)
const previewing = ref(false)
const previewHtml = ref('')
const savedOk = ref(false)

async function load() {
  error.value = ''
  try {
    const res: any = await $fetch('/api/commercial/email-templates')
    templates.value = res.data ?? res ?? []
  } catch (e: any) {
    error.value = mapError(e)
  }
}

function select(tpl: any) {
  edit.value = {
    id: tpl.id,
    slug: tpl.slug,
    name: tpl.name,
    category: tpl.category,
    subject: tpl.subject,
    bodyHtml: tpl.bodyHtml,
    isActive: tpl.isActive,
  }
  previewHtml.value = ''
  savedOk.value = false
}

async function save() {
  if (!edit.value) return
  saving.value = true
  error.value = ''
  try {
    await $fetch(`/api/commercial/email-templates/${edit.value.id}`, {
      method: 'PATCH',
      body: {
        name: edit.value.name,
        category: edit.value.category,
        subject: edit.value.subject,
        bodyHtml: edit.value.bodyHtml,
        isActive: edit.value.isActive,
      },
    })
    await load()
    savedOk.value = true
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    saving.value = false
  }
}

async function preview() {
  if (!edit.value) return
  previewing.value = true
  error.value = ''
  try {
    const res: any = await $fetch(`/api/commercial/email-templates/${edit.value.id}/preview`, {
      method: 'POST',
      body: {
        subject: edit.value.subject,
        bodyHtml: edit.value.bodyHtml,
      },
    })
    const data = res.data ?? res
    previewHtml.value = data.fullHtml || data.bodyHtml || ''
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    previewing.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.pf-mail-layout {
  display: grid;
  gap: 1.25rem;
}
@media (min-width: 960px) {
  .pf-mail-layout {
    grid-template-columns: 1fr 1.2fr;
    align-items: start;
  }
}
.pf-mail-body {
  font-family: var(--pf-font-mono, ui-monospace, monospace);
  font-size: 0.85rem;
  min-height: 220px;
}
.pf-mail-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.pf-mail-preview {
  width: 100%;
  min-height: 420px;
  margin-top: 1rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: #fff;
}
.pro-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.75rem 0;
}
</style>
