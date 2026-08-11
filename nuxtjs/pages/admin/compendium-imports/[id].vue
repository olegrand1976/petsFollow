<template>
  <div data-testid="admin-compendium-import-detail">
    <ProPageHeader
      :title="$t('admin.compendium.detailTitle')"
      :subtitle="job ? `${job.filename} · p.${job.pageStart}–${job.pageEnd}` : ''"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
        <NuxtLink to="/admin/compendium-imports" class="pro-link">{{ $t('admin.compendium.back') }}</NuxtLink>
      </template>
    </ProPageHeader>

    <div v-if="loading" class="pro-hint">{{ $t('common.loading') }}</div>
    <template v-else-if="job">
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi :value="`${job.extractPct ?? 0}%`" :label="$t('admin.compendium.kpiExtract')" />
        <ProKpi :value="`${job.reviewPct ?? 0}%`" :label="$t('admin.compendium.kpiReview')" />
        <ProKpi :value="totals.ready" :label="$t('admin.compendium.kpiReady')" />
        <ProKpi :value="job.upsertedCount" :label="$t('admin.compendium.kpiUpserted')" />
      </div>

      <ProCard class="pro-mb-lg" data-testid="admin-compendium-progress">
        <h3 class="pro-mb-md">{{ $t('admin.compendium.progressTitle') }}</h3>
        <p class="pro-hint">{{ $t('admin.compendium.progressExtract', { pct: job.extractPct ?? 0, done: job.extractDone, total: job.extractTotal }) }}</p>
        <div class="compendium-bar" role="progressbar" :aria-valuenow="job.extractPct ?? 0" aria-valuemin="0" aria-valuemax="100">
          <div class="compendium-bar__fill" :style="{ width: `${job.extractPct ?? 0}%` }" />
        </div>
        <p class="pro-hint pro-mt-md">{{ $t('admin.compendium.progressReview', { pct: job.reviewPct ?? 0, done: job.reviewedCount, total: job.rowCount }) }}</p>
        <div class="compendium-bar" role="progressbar" :aria-valuenow="job.reviewPct ?? 0" aria-valuemin="0" aria-valuemax="100">
          <div class="compendium-bar__fill compendium-bar__fill--review" :style="{ width: `${job.reviewPct ?? 0}%` }" />
        </div>
        <p v-if="job.errorMessage" class="pro-hint pro-hint--error pro-mt-md">{{ job.errorMessage }}</p>
        <div class="pro-flex-gap pro-mt-md">
          <ProButton
            v-if="job.status === 'uploaded' || job.status === 'failed' || job.status === 'extracting'"
            test-id="admin-compendium-extract"
            :disabled="busy"
            @click="startExtract(false)"
          >
            {{ extractPrimaryLabel }}
          </ProButton>
          <ProButton
            v-if="canForceRestartExtract"
            variant="secondary"
            test-id="admin-compendium-extract-restart"
            :disabled="busy"
            @click="startExtract(true)"
          >
            {{ $t('admin.compendium.restartExtract') }}
          </ProButton>
          <ProBadge :variant="statusVariant(job.status)">{{ statusLabel(job.status) }}</ProBadge>
        </div>
        <p v-if="canResumeExtract" class="pro-hint pro-mt-md" data-testid="admin-compendium-resume-hint">
          {{ $t('admin.compendium.resumeHint', { done: job.extractDone, total: job.extractTotal }) }}
        </p>
      </ProCard>

      <ProCard class="pro-mb-lg" data-testid="admin-compendium-review-toolbar">
        <div class="compendium-totals pro-mb-md" data-testid="admin-compendium-totals">
          <span class="compendium-chip">{{ $t('admin.compendium.totalToReview') }} · <strong>{{ totals.toReview }}</strong></span>
          <span class="compendium-chip compendium-chip--ok">{{ $t('admin.compendium.totalReady') }} · <strong>{{ totals.ready }}</strong></span>
          <span class="compendium-chip">{{ $t('admin.compendium.totalExcluded') }} · <strong>{{ totals.excluded }}</strong></span>
          <span class="compendium-chip">{{ $t('admin.compendium.totalLoaded') }} · <strong>{{ totals.loaded }}</strong></span>
        </div>
          <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.previewHint') }}</p>
          <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.filtersScopeHint') }}</p>
        <p v-if="errorRowCount > 0" class="pro-hint pro-mb-md" data-testid="admin-compendium-manual-hint">
          {{ $t('admin.compendium.manualHint') }}
        </p>
        <p
          v-if="refCatalogCount === 0 && job.status === 'extracted'"
          class="pro-hint pro-hint--error pro-mb-md"
          data-testid="admin-compendium-catalog-empty"
        >
          {{ $t('admin.compendium.catalogEmpty') }}
        </p>
        <h4 class="pro-mb-sm">{{ $t('admin.compendium.reviewFiltersTitle') }}</h4>
        <div class="compendium-filters pro-mb-md" data-testid="admin-compendium-filters">
          <input
            v-model="filterQ"
            type="search"
            class="pro-input compendium-filters__search"
            data-testid="admin-compendium-filter-q"
            :placeholder="$t('admin.compendium.filterSearch')"
            @input="resetReviewPage"
          >
          <label class="compendium-filters__status">
            <span class="pro-hint">{{ $t('admin.compendium.filterStatus') }}</span>
            <select
              v-model="filterStatus"
              class="pro-input"
              data-testid="admin-compendium-filter-status"
              @change="resetReviewPage"
            >
              <option value="all">{{ $t('admin.compendium.filterStatusAll') }}</option>
              <option value="pending">{{ $t('admin.compendium.filterStatusPending') }}</option>
              <option value="error">{{ $t('admin.compendium.filterStatusError') }}</option>
            </select>
          </label>
          <label class="compendium-filters__check">
            <input
              v-model="filterHasSuggested"
              type="checkbox"
              data-testid="admin-compendium-filter-suggested"
              @change="resetReviewPage"
            >
            {{ $t('admin.compendium.filterHasSuggested') }}
          </label>
        </div>
        <div v-if="pendingCount > 0 && job.status === 'extracted'" class="pro-flex-gap">
          <ProButton
            test-id="admin-compendium-confirm-ready"
            variant="secondary"
            :disabled="busy"
            @click="confirmReady"
          >
            {{ $t('admin.compendium.confirmReady', { count: pendingCount }) }}
          </ProButton>
        </div>
      </ProCard>

      <div class="compendium-review pro-mb-lg" data-testid="admin-compendium-preview">
        <div class="compendium-dual">
          <ProCard class="compendium-panel" data-testid="admin-compendium-panel-review">
            <div class="compendium-panel__head">
              <h3>{{ $t('admin.compendium.panelReview') }}</h3>
              <span class="pro-hint" data-testid="admin-compendium-review-range">
                {{ $t('admin.compendium.pageRange', { from: reviewPage.from, to: reviewPage.to, total: reviewPage.total }) }}
              </span>
            </div>
            <ProTable :empty="!reviewPage.items.length" :empty-title="$t('admin.compendium.emptyReview')">
              <thead>
                <tr>
                  <th>#</th>
                  <th>{{ $t('admin.compendium.colPage') }}</th>
                  <th>{{ $t('admin.compendium.colCnk') }}</th>
                  <th>{{ $t('admin.compendium.colSuggestedCnk') }}</th>
                  <th>{{ $t('admin.compendium.colName') }}</th>
                  <th>{{ $t('admin.compendium.colLab') }}</th>
                  <th>{{ $t('admin.compendium.colSubstance') }}</th>
                  <th>{{ $t('admin.compendium.colStrength') }}</th>
                  <th>{{ $t('admin.compendium.colAtc') }}</th>
                  <th>{{ $t('admin.compendium.colForm') }}</th>
                  <th>{{ $t('admin.compendium.colPack') }}</th>
                  <th>{{ $t('admin.compendium.colStatus') }}</th>
                  <th>{{ $t('admin.compendium.colActions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in reviewPage.items"
                  :key="row.id"
                  :class="{
                    'compendium-row--needs-manual': row.status === 'error',
                    'compendium-row--focused': focusedRowId === row.id,
                  }"
                  :data-testid="row.status === 'error' ? 'admin-compendium-row-error' : undefined"
                  @click="focusRow(row)"
                >
                  <td>{{ row.rowNumber }}</td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model.number="row.sourcePage"
                      type="number"
                      min="1"
                      class="pro-input compendium-page-input"
                      data-testid="admin-compendium-row-page"
                      :title="$t('admin.compendium.pageHint', { page: row.sourcePage ?? '—' })"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @focus="focusRow(row)"
                      @change="onSourcePageChange(row)"
                    >
                    <button
                      v-else
                      type="button"
                      class="compendium-page compendium-page-btn"
                      data-testid="admin-compendium-row-page"
                      @click.stop="focusRow(row)"
                    >{{ row.sourcePage ?? '—' }}</button>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.cnk"
                      class="pro-input"
                      :placeholder="$t('admin.compendium.colCnk')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { cnk: row.cnk })"
                    >
                    <span v-else>{{ row.cnk }}</span>
                  </td>
                  <td>
                    <template v-if="rowEditable(row) && candidates(row).length">
                      <select
                        class="pro-input"
                        data-testid="admin-compendium-cnk-candidates"
                        :value="row.cnk || row.suggestedCnk || ''"
                        :disabled="isRowBusy(row)"
                        @click.stop
                        @change="onPickCandidate(row, ($event.target as HTMLSelectElement).value)"
                      >
                        <option value="">{{ $t('admin.compendium.pickCnk') }}</option>
                        <option
                          v-for="c in candidates(row)"
                          :key="c.cnk"
                          :value="c.cnk"
                        >
                          {{ c.cnk }} — {{ c.name }} ({{ Math.round((c.score || 0) * 100) }}%)
                        </option>
                      </select>
                    </template>
                    <span v-else-if="row.suggestedCnk">{{ row.suggestedCnk }}</span>
                    <span v-else>—</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.name"
                      class="pro-input"
                      :placeholder="$t('admin.compendium.colName')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { name: row.name })"
                    >
                    <span v-else>{{ row.name }}</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.manufacturer"
                      class="pro-input"
                      :placeholder="$t('admin.compendium.colLab')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { manufacturer: row.manufacturer })"
                    >
                    <span v-else>{{ row.manufacturer || '—' }}</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.activeSubstance"
                      class="pro-input"
                      :placeholder="$t('admin.compendium.colSubstance')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { activeSubstance: row.activeSubstance })"
                    >
                    <span v-else>{{ row.activeSubstance || '—' }}</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.strength"
                      class="pro-input"
                      data-testid="admin-compendium-row-strength"
                      :placeholder="$t('admin.compendium.colStrength')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { strength: row.strength })"
                    >
                    <span v-else>{{ row.strength || '—' }}</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.atcCode"
                      class="pro-input"
                      data-testid="admin-compendium-row-atc"
                      :placeholder="$t('admin.compendium.colAtc')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { atcCode: row.atcCode })"
                    >
                    <span v-else>{{ row.atcCode || '—' }}</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.pharmaceuticalForm"
                      class="pro-input"
                      :placeholder="$t('admin.compendium.colForm')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { pharmaceuticalForm: row.pharmaceuticalForm })"
                    >
                    <span v-else>{{ row.pharmaceuticalForm || '—' }}</span>
                  </td>
                  <td>
                    <input
                      v-if="rowEditable(row)"
                      v-model="row.packSize"
                      class="pro-input"
                      :placeholder="$t('admin.compendium.colPack')"
                      :disabled="isRowBusy(row)"
                      @click.stop
                      @change="patchRow(row, { packSize: row.packSize })"
                    >
                    <span v-else>{{ row.packSize || '—' }}</span>
                  </td>
                  <td>
                    <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                    <span v-if="row.errorCode" class="pro-hint"> {{ row.errorCode }}</span>
                    <p v-if="row.status === 'error' && row.sourcePage" class="pro-hint">
                      {{ $t('admin.compendium.fixFromPage', { page: row.sourcePage }) }}
                    </p>
                  </td>
                  <td @click.stop>
                    <ProButton
                      v-if="row.status === 'pending' || row.status === 'error'"
                      variant="ghost"
                      test-id="admin-compendium-lookup-cnk"
                      :disabled="busy || isRowBusy(row) || !row.name"
                      @click="lookupCnk(row)"
                    >
                      {{ $t('admin.compendium.lookupCnk') }}
                    </ProButton>
                    <ProButton
                      v-if="(row.status === 'pending' || row.status === 'error') && row.suggestedCnk && !row.cnk"
                      variant="ghost"
                      :disabled="isRowBusy(row)"
                      @click="patchRow(row, { cnk: row.suggestedCnk })"
                    >
                      {{ $t('admin.compendium.acceptSuggested') }}
                    </ProButton>
                    <ProButton
                      v-if="canConfirmRow(row)"
                      variant="ghost"
                      test-id="admin-compendium-confirm-row"
                      :disabled="busy || isRowBusy(row)"
                      @click="patchRow(row, { name: row.name, cnk: row.cnk })"
                    >
                      {{ $t('admin.compendium.confirmRow') }}
                    </ProButton>
                    <ProButton
                      v-if="rowEditable(row)"
                      variant="ghost"
                      :disabled="isRowBusy(row)"
                      @click="patchRow(row, { excluded: true })"
                    >
                      {{ $t('admin.compendium.exclude') }}
                    </ProButton>
                    <p v-if="lookupMsg[row.id]" class="pro-hint" data-testid="admin-compendium-lookup-msg">{{ lookupMsg[row.id] }}</p>
                  </td>
                </tr>
              </tbody>
            </ProTable>
            <div class="compendium-pager pro-mt-md" data-testid="admin-compendium-review-pager">
              <ProButton variant="ghost" :disabled="reviewPage.page <= 1" @click="reviewPageNum--">
                {{ $t('admin.compendium.prevPage') }}
              </ProButton>
              <ProButton variant="ghost" :disabled="reviewPage.page >= reviewPage.pageCount" @click="reviewPageNum++">
                {{ $t('admin.compendium.nextPage') }}
              </ProButton>
            </div>
          </ProCard>

          <ProCard class="compendium-panel" data-testid="admin-compendium-panel-ready">
            <div class="compendium-panel__head">
              <h3>{{ $t('admin.compendium.panelReady') }}</h3>
              <span class="pro-hint" data-testid="admin-compendium-ready-range">
                {{ $t('admin.compendium.pageRange', { from: readyPage.from, to: readyPage.to, total: readyPage.total }) }}
              </span>
            </div>
            <ProTable :empty="!readyPage.items.length" :empty-title="$t('admin.compendium.emptyReady')">
              <thead>
                <tr>
                  <th>#</th>
                  <th>{{ $t('admin.compendium.colPage') }}</th>
                  <th>{{ $t('admin.compendium.colCnk') }}</th>
                  <th>{{ $t('admin.compendium.colName') }}</th>
                  <th>{{ $t('admin.compendium.colLab') }}</th>
                  <th>{{ $t('admin.compendium.colAtc') }}</th>
                  <th>{{ $t('admin.compendium.colStatus') }}</th>
                  <th>{{ $t('admin.compendium.colActions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in readyPage.items"
                  :key="row.id"
                  :class="{ 'compendium-row--focused': focusedRowId === row.id }"
                  data-testid="admin-compendium-row-ready"
                  @click="focusRow(row)"
                >
                  <td>{{ row.rowNumber }}</td>
                  <td>
                    <button
                      type="button"
                      class="compendium-page compendium-page-btn"
                      @click.stop="focusRow(row)"
                    >{{ row.sourcePage ?? '—' }}</button>
                  </td>
                  <td>{{ row.cnk }}</td>
                  <td>{{ row.name }}</td>
                  <td>{{ row.manufacturer || '—' }}</td>
                  <td>{{ row.atcCode || '—' }}</td>
                  <td>
                    <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                  </td>
                  <td @click.stop>
                    <ProButton
                      variant="ghost"
                      :disabled="isRowBusy(row)"
                      @click="patchRow(row, { excluded: true })"
                    >
                      {{ $t('admin.compendium.exclude') }}
                    </ProButton>
                  </td>
                </tr>
              </tbody>
            </ProTable>
            <div class="compendium-pager pro-mt-md" data-testid="admin-compendium-ready-pager">
              <ProButton variant="ghost" :disabled="readyPage.page <= 1" @click="readyPageNum--">
                {{ $t('admin.compendium.prevPage') }}
              </ProButton>
              <ProButton variant="ghost" :disabled="readyPage.page >= readyPage.pageCount" @click="readyPageNum++">
                {{ $t('admin.compendium.nextPage') }}
              </ProButton>
            </div>
          </ProCard>

          <ProCard class="compendium-panel" data-testid="admin-compendium-panel-excluded">
            <div class="compendium-panel__head">
              <h3>{{ $t('admin.compendium.panelExcluded') }}</h3>
              <span class="pro-hint" data-testid="admin-compendium-excluded-range">
                {{ $t('admin.compendium.pageRange', { from: excludedPage.from, to: excludedPage.to, total: excludedPage.total }) }}
              </span>
            </div>
            <ProTable :empty="!excludedPage.items.length" :empty-title="$t('admin.compendium.emptyExcluded')">
              <thead>
                <tr>
                  <th>#</th>
                  <th>{{ $t('admin.compendium.colPage') }}</th>
                  <th>{{ $t('admin.compendium.colCnk') }}</th>
                  <th>{{ $t('admin.compendium.colName') }}</th>
                  <th>{{ $t('admin.compendium.colStatus') }}</th>
                  <th>{{ $t('admin.compendium.colActions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in excludedPage.items"
                  :key="row.id"
                  :class="{ 'compendium-row--focused': focusedRowId === row.id }"
                  data-testid="admin-compendium-row-excluded"
                  @click="focusRow(row)"
                >
                  <td>{{ row.rowNumber }}</td>
                  <td>
                    <button
                      type="button"
                      class="compendium-page compendium-page-btn"
                      @click.stop="focusRow(row)"
                    >{{ row.sourcePage ?? '—' }}</button>
                  </td>
                  <td>{{ row.cnk || '—' }}</td>
                  <td>{{ row.name || '—' }}</td>
                  <td>
                    <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                  </td>
                  <td @click.stop>
                    <ProButton
                      variant="ghost"
                      test-id="admin-compendium-include-row"
                      :disabled="isRowBusy(row)"
                      @click="patchRow(row, { excluded: false })"
                    >
                      {{ $t('admin.compendium.include') }}
                    </ProButton>
                  </td>
                </tr>
              </tbody>
            </ProTable>
            <div class="compendium-pager pro-mt-md" data-testid="admin-compendium-excluded-pager">
              <ProButton variant="ghost" :disabled="excludedPage.page <= 1" @click="excludedPageNum--">
                {{ $t('admin.compendium.prevPage') }}
              </ProButton>
              <ProButton variant="ghost" :disabled="excludedPage.page >= excludedPage.pageCount" @click="excludedPageNum++">
                {{ $t('admin.compendium.nextPage') }}
              </ProButton>
            </div>
          </ProCard>
        </div>

        <ProCard class="compendium-review__viewer" data-testid="admin-compendium-pdf-viewer">
          <h3 class="pro-mb-md">{{ $t('admin.compendium.viewerTitle') }}</h3>
          <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.viewerHint') }}</p>
          <div class="pro-flex-gap pro-mb-md">
            <label class="compendium-viewer-page">
              <span class="pro-hint">{{ $t('admin.compendium.colPage') }}</span>
              <input
                v-model.number="viewerPage"
                type="number"
                min="1"
                class="pro-input compendium-page-input"
                data-testid="admin-compendium-viewer-page"
                @change="applyViewerPage"
                @keyup.enter="applyViewerPage"
              >
            </label>
            <ProButton
              variant="secondary"
              test-id="admin-compendium-open-pdf"
              @click="openPdfTab"
            >
              {{ $t('admin.compendium.openPdfTab') }}
            </ProButton>
          </div>
          <iframe
            v-if="pdfIframeSrc"
            class="compendium-pdf-frame"
            :src="pdfIframeSrc"
            :title="$t('admin.compendium.viewerTitle')"
            data-testid="admin-compendium-pdf-frame"
          />
        </ProCard>
      </div>

      <ProCard data-testid="admin-compendium-commit">
        <h3 class="pro-mb-md">{{ $t('admin.compendium.commitTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.commitHint') }}</p>
        <ProButton
          test-id="admin-compendium-commit"
          :disabled="busy || job.status !== 'extracted' || !totals.ready"
          @click="commit"
        >
          {{ $t('admin.compendium.commit') }}
        </ProButton>
        <p v-if="commitMsg" class="pro-hint pro-mt-md" data-testid="admin-compendium-commit-msg">{{ commitMsg }}</p>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import {
  COMPENDIUM_REVIEW_PAGE_SIZE,
  compendiumReviewTotals,
  filterCompendiumReviewRows,
  isExcludedQueueRow,
  isReadyQueueRow,
  isReviewQueueRow,
  paginateRows,
  type CompendiumReviewStatusFilter,
} from '~/utils/compendium-review'

definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const route = useRoute()
const id = computed(() => String(route.params.id))
const loading = ref(true)
const busy = ref(false)
const job = ref<any>(null)
const rows = ref<any[]>([])
const refCatalogCount = ref(0)
const commitMsg = ref('')
const lookingUpId = ref('')
const patchingId = ref('')
const lookupMsg = ref<Record<string, string>>({})
const focusedRowId = ref('')
const viewerPage = ref(0)
/** Page applied to the iframe (not every keystroke — avoids re-download of large PDFs). */
const appliedViewerPage = ref(0)
const filterQ = ref('')
const filterStatus = ref<CompendiumReviewStatusFilter>('all')
const filterHasSuggested = ref(false)
const reviewPageNum = ref(1)
const readyPageNum = ref(1)
const excludedPageNum = ref(1)
let pollTimer: ReturnType<typeof setInterval> | null = null

const pendingCount = computed(() => rows.value.filter(r => canConfirmRow(r)).length)
const errorRowCount = computed(() => rows.value.filter(r => r.status === 'error').length)
const totals = computed(() => compendiumReviewTotals(rows.value))

const filteredReviewRows = computed(() => filterCompendiumReviewRows(
  rows.value.filter(isReviewQueueRow),
  {
    q: filterQ.value,
    status: filterStatus.value,
    hasSuggested: filterHasSuggested.value,
  },
))
const readyRows = computed(() => rows.value.filter(isReadyQueueRow))
const excludedRows = computed(() => rows.value.filter(isExcludedQueueRow))
const reviewPage = computed(() => paginateRows(filteredReviewRows.value, reviewPageNum.value, COMPENDIUM_REVIEW_PAGE_SIZE))
const readyPage = computed(() => paginateRows(readyRows.value, readyPageNum.value, COMPENDIUM_REVIEW_PAGE_SIZE))
const excludedPage = computed(() => paginateRows(excludedRows.value, excludedPageNum.value, COMPENDIUM_REVIEW_PAGE_SIZE))

watch(reviewPage, (p) => {
  if (reviewPageNum.value !== p.page) reviewPageNum.value = p.page
})
watch(readyPage, (p) => {
  if (readyPageNum.value !== p.page) readyPageNum.value = p.page
})
watch(excludedPage, (p) => {
  if (excludedPageNum.value !== p.page) excludedPageNum.value = p.page
})

const pdfBaseUrl = computed(() => `/api/admin/compendium-imports/${id.value}/pdf`)
const pdfIframeSrc = computed(() => {
  if (!job.value) return ''
  const page = Number(appliedViewerPage.value) > 0
    ? Number(appliedViewerPage.value)
    : (Number(job.value.pageStart) || 1)
  return `${pdfBaseUrl.value}#page=${page}`
})
const canResumeExtract = computed(() => {
  if (!job.value) return false
  if (job.value.status !== 'failed' && job.value.status !== 'extracting') return false
  const done = Number(job.value.extractDone || 0)
  const total = Number(job.value.extractTotal || 0)
  return done > 0 && total > done
})
const canForceRestartExtract = computed(() => canResumeExtract.value)
const extractPrimaryLabel = computed(() => {
  if (!job.value) return ''
  if (job.value.status === 'uploaded') return t('admin.compendium.startExtract')
  if (canResumeExtract.value) return t('admin.compendium.resumeExtract')
  return t('admin.compendium.retryExtract')
})

function resetReviewPage () {
  reviewPageNum.value = 1
}

function rowEditable (row: any) {
  return row?.status !== 'upserted' && row?.status !== 'excluded'
}

function isRowBusy (row: any) {
  return patchingId.value === row?.id || lookingUpId.value === row?.id
}

function canConfirmRow (row: any) {
  if (!rowEditable(row)) return false
  if (row.status !== 'pending' && row.status !== 'error') return false
  return Boolean(String(row.cnk || '').trim() && String(row.name || '').trim())
}

function setViewerPage (page: number) {
  const n = Math.trunc(Number(page))
  if (!Number.isFinite(n) || n < 1) return
  viewerPage.value = n
  appliedViewerPage.value = n
}

function applyViewerPage () {
  setViewerPage(viewerPage.value)
}

function focusRow (row: any) {
  if (!row?.id) return
  focusedRowId.value = row.id
  const page = Number(row.sourcePage)
  if (page > 0) setViewerPage(page)
  else if (job.value?.pageStart) setViewerPage(Number(job.value.pageStart) || 1)
}

function onSourcePageChange (row: any) {
  focusRow(row)
  void patchSourcePage(row)
}

function openPdfTab () {
  const page = Number(appliedViewerPage.value) > 0
    ? Number(appliedViewerPage.value)
    : (Number(viewerPage.value) > 0 ? Number(viewerPage.value) : (Number(job.value?.pageStart) || 1))
  window.open(`${pdfBaseUrl.value}#page=${page}`, '_blank', 'noopener,noreferrer')
}

function candidates (row: any): Array<{ cnk: string; name: string; score?: number }> {
  const raw = row?.matchCandidates
  if (Array.isArray(raw)) return raw
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

function onPickCandidate (row: any, cnk: string) {
  if (!cnk) return
  row.cnk = cnk
  void patchRow(row, { cnk })
}

function statusVariant (status: string) {
  switch (status) {
    case 'completed': case 'ready': case 'upserted': return 'success'
    case 'failed': case 'error': return 'danger'
    case 'extracting': case 'committing': case 'pending': return 'warning'
    default: return 'neutral'
  }
}
function rowStatusVariant (status: string) {
  return statusVariant(status)
}
function statusLabel (status: string) {
  return t(`admin.compendium.status.${status}`, status)
}

async function load () {
  const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}`)
  const data = res?.data ?? res
  job.value = data?.job ?? data
  rows.value = data?.rows ?? []
  refCatalogCount.value = Number(data?.refCatalogCount ?? 0)
  loading.value = false
  if (!viewerPage.value && job.value?.pageStart) {
    setViewerPage(Number(job.value.pageStart) || 1)
  }
  if (job.value?.status === 'extracting') {
    startPoll()
  } else {
    stopPoll()
  }
}

function startPoll () {
  if (pollTimer) return
  pollTimer = setInterval(() => { void load() }, 1500)
}
function stopPoll () {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function startExtract (forceRestart = false) {
  busy.value = true
  try {
    const q = forceRestart ? '?restart=1' : ''
    await $fetch(`/api/admin/compendium-imports/${id.value}/extract${q}`, { method: 'POST' })
    startPoll()
    await load()
  } finally {
    busy.value = false
  }
}

async function patchSourcePage (row: any) {
  const n = Number(row.sourcePage)
  if (!Number.isFinite(n) || n < 1) {
    row.sourcePage = null
    // 0 clears source_page server-side (*int null vs omit is ambiguous).
    await patchRow(row, { sourcePage: 0 })
    return
  }
  row.sourcePage = Math.trunc(n)
  await patchRow(row, { sourcePage: row.sourcePage })
}

async function patchRow (row: any, body: Record<string, unknown>) {
  patchingId.value = row.id
  try {
    const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}/rows/${row.id}`, {
      method: 'PATCH',
      body,
    })
    const updated = res?.data ?? res
    const idx = rows.value.findIndex(r => r.id === row.id)
    if (idx >= 0) rows.value[idx] = { ...rows.value[idx], ...updated }
    await load()
  } finally {
    patchingId.value = ''
  }
}

async function lookupCnk (row: any) {
  if (!row?.id || !row.name) return
  lookingUpId.value = row.id
  lookupMsg.value = { ...lookupMsg.value, [row.id]: '' }
  busy.value = true
  try {
    const body: Record<string, string> = {}
    if (row.cnk) body.cnk = String(row.cnk).trim()
    const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}/rows/${row.id}/lookup-cnk`, {
      method: 'POST',
      body,
    })
    const data = res?.data ?? res
    const updated = data?.row ?? data
    const idx = rows.value.findIndex(r => r.id === row.id)
    if (idx >= 0 && updated) rows.value[idx] = { ...rows.value[idx], ...updated }
    if (typeof data?.cnkFound === 'boolean') {
      lookupMsg.value = {
        ...lookupMsg.value,
        [row.id]: data.cnkFound
          ? t('admin.compendium.lookupCnkFound')
          : t('admin.compendium.lookupCnkNotFound'),
      }
    } else {
      lookupMsg.value = {
        ...lookupMsg.value,
        [row.id]: t('admin.compendium.lookupCnkDone'),
      }
    }
    await load()
  } catch (e: any) {
    lookupMsg.value = {
      ...lookupMsg.value,
      [row.id]: e?.data?.error?.message ?? e?.statusMessage ?? t('admin.compendium.lookupCnkFailed'),
    }
  } finally {
    lookingUpId.value = ''
    busy.value = false
  }
}

async function confirmReady () {
  busy.value = true
  try {
    await $fetch(`/api/admin/compendium-imports/${id.value}/confirm-ready`, { method: 'POST' })
    readyPageNum.value = 1
    await load()
  } finally {
    busy.value = false
  }
}

async function commit () {
  if (!confirm(t('admin.compendium.commitConfirm'))) return
  busy.value = true
  commitMsg.value = ''
  try {
    const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}/commit`, { method: 'POST' })
    const result = res?.data?.result ?? res?.result
    commitMsg.value = t('admin.compendium.commitOk', {
      upserted: result?.upserted ?? 0,
      skipped: result?.skipped ?? 0,
    })
    await load()
  } catch (e: any) {
    commitMsg.value = e?.data?.error?.message ?? t('admin.compendium.commitFailed')
  } finally {
    busy.value = false
  }
}

onMounted(() => { void load() })
onBeforeUnmount(() => stopPoll())
</script>

<style scoped>
.compendium-bar {
  height: 8px;
  border-radius: 4px;
  background: var(--pf-vet-border);
  overflow: hidden;
}
.compendium-bar__fill {
  height: 100%;
  background: var(--pf-vet-accent);
  transition: width 0.3s ease;
}
.compendium-bar__fill--review {
  background: var(--pf-vet-primary);
}
.compendium-totals {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.compendium-chip {
  display: inline-flex;
  align-items: baseline;
  gap: 0.25rem;
  padding: 0.35rem 0.65rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 6px;
  background: var(--pf-vet-surface, #fff);
  font-size: 0.9rem;
}
.compendium-chip--ok {
  border-color: color-mix(in srgb, var(--pf-vet-primary) 35%, var(--pf-vet-border));
}
.compendium-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: end;
}
.compendium-filters__search {
  min-width: min(100%, 18rem);
  flex: 1 1 14rem;
}
.compendium-filters__status {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.compendium-filters__check {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding-bottom: 0.35rem;
}
.compendium-review {
  display: grid;
  gap: 1rem;
  align-items: start;
}
.compendium-dual {
  display: grid;
  gap: 1rem;
}
.compendium-panel__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.compendium-panel__head h3 {
  margin: 0;
}
.compendium-pager {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
@media (min-width: 1100px) {
  .compendium-review {
    grid-template-columns: minmax(0, 1.45fr) minmax(280px, 0.9fr);
  }
  .compendium-review__viewer {
    position: sticky;
    top: 1rem;
  }
}
.compendium-viewer-page {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}
.compendium-pdf-frame {
  width: 100%;
  min-height: 70vh;
  border: 1px solid var(--pf-vet-border);
  border-radius: 6px;
  background: var(--pf-vet-surface, #fff);
}
.compendium-page {
  font-family: var(--pf-font-mono, ui-monospace, monospace);
  font-weight: 600;
}
.compendium-page-btn {
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.compendium-page-input {
  width: 5.5rem;
  font-family: var(--pf-font-mono, ui-monospace, monospace);
  font-weight: 600;
}
.compendium-row--needs-manual td {
  background: color-mix(in srgb, var(--pf-vet-alert) 8%, transparent);
}
.compendium-row--focused td {
  outline: 1px solid color-mix(in srgb, var(--pf-vet-primary) 35%, transparent);
}
</style>
