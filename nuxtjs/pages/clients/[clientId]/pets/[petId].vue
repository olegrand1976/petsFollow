<template>
  <div data-testid="pet-detail-page">
    <nav class="pro-breadcrumb" :aria-label="$t('common.breadcrumb')">
      <NuxtLink to="/pets">{{ $t('nav.pets') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <NuxtLink to="/clients">{{ $t('nav.clients') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <NuxtLink :to="`/clients/${clientId}`">{{ $t('common.profile') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ pet?.name || $t('clients.pet.title') }}</span>
    </nav>
    <ProPageHeader
      :title="pet?.name || $t('clients.pet.title')"
      :subtitle="petSubtitle"
    >
      <template #actions>
        <div class="pro-pet-header-actions">
          <ProBadge v-if="isPrimaryPractice" variant="success">{{ $t('clients.pet.primaryBadge') }}</ProBadge>
          <ProButton
            v-if="canWriteClinical"
            variant="primary"
            test-id="pet-new-consultation"
            @click="openConsultation"
          >
            <ProIcon name="medical_services" :size="18" />
            {{ $t('clients.consultation.open') }}
          </ProButton>
          <ProButton
            v-if="canMessage"
            variant="secondary"
            test-id="pet-open-messages"
            :disabled="messagingBusy"
            @click="openMessaging"
          >
            <ProIcon name="chat" :size="18" />
            {{ $t('clients.pet.openMessages') }}
          </ProButton>
          <ProAvatar :src="pet?.photoUrl" :name="pet?.name || ''" size="lg" />
        </div>
      </template>
    </ProPageHeader>
    <p v-if="pageError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ pageError }}</p>
    <p v-if="messagingError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ messagingError }}</p>

    <div class="pro-grid-kpi" data-testid="pet-kpi-strip">
      <ProKpi
        icon="favorite"
        :value="kpiLastBpm"
        :label="$t('clients.pet.kpi.lastBpm')"
        :variant="recentAlert ? 'alert' : 'default'"
      />
      <ProKpi
        icon="warning"
        :value="kpiAlertCount"
        :label="$t('clients.pet.kpi.alerts')"
        :variant="kpiAlertCount > 0 ? 'alert' : 'default'"
      />
      <ProKpi
        icon="monitor_weight"
        :value="kpiLastWeight"
        :label="$t('clients.pet.kpi.lastWeight')"
        :trend="kpiWeightDelta || undefined"
      />
      <ProKpi
        icon="monitor_heart"
        :value="kpiLastBp"
        :label="$t('clients.pet.kpi.lastBp')"
      />
      <ProKpi
        icon="medical_services"
        :value="kpiOverdueCare"
        :label="$t('clients.pet.kpi.overdueCare')"
        :variant="kpiOverdueCare > 0 ? 'alert' : 'default'"
      />
      <ProKpi
        icon="event"
        :value="kpiNextVisit"
        :label="$t('clients.pet.kpi.nextVisit')"
      />
    </div>

    <div v-if="recentAlert" class="pro-alert-banner" role="status" data-testid="pet-alert-banner">
      <ProBadge variant="danger">{{ $t('clients.pet.alert') }}</ProBadge>
      <span>{{ $t('clients.pet.recentAlertBanner') }}</span>
    </div>

    <ProSectionTabs
      v-model="activeTab"
      :tabs="petTabs"
      :aria-label="$t('clients.pet.tabsLabel')"
    />

    <div
      v-show="activeTab === 'overview'"
      role="tabpanel"
      aria-labelledby="tab-overview"
      data-testid="pet-tab-overview"
    >
      <ProCard :title="$t('clients.pet.photoTitle')" class="pro-settings-card">
        <ProAvatarUpload
          v-model="petPhotoUrl"
          :name="pet?.name || ''"
          :upload-url="`/api/pets/${petId}/photo`"
          :label="$t('clients.pet.photoChange')"
          :hint="$t('clients.pet.photoHint')"
          :disabled="!canWriteClinical"
          @uploaded="onPetPhotoUploaded"
        />
      </ProCard>
      <ProCard v-if="pet" :title="$t('clients.pet.summaryTitle')" class="pro-mb-lg" data-testid="pet-medical-data">
        <dl class="pro-pet-summary">
          <div>
            <dt>{{ $t('pets.columnSpecies') }}</dt>
            <dd>{{ speciesLabel(pet.species) }}</dd>
          </div>
          <div>
            <dt>{{ $t('pets.columnBreed') }}</dt>
            <dd>{{ pet.breed || $t('common.unknownBreed') }}</dd>
          </div>
          <div>
            <dt>{{ $t('pets.columnBirthDate') }}</dt>
            <dd data-testid="pet-birth-date">{{ formatPetDay(pet.birthDate) }}</dd>
          </div>
          <div v-if="pet.weightKg != null">
            <dt>{{ $t('clients.pet.columnWeight') }}</dt>
            <dd>{{ pet.weightKg }} kg</dd>
          </div>
          <div>
            <dt>{{ $t('clients.pet.microchip') }}</dt>
            <dd data-testid="pet-microchip">{{ pet.microchipNumber || $t('common.dash') }}</dd>
          </div>
          <div>
            <dt>{{ $t('clients.pet.healthBookNumber') }}</dt>
            <dd data-testid="pet-health-book-number">{{ pet.healthBookNumber || $t('common.dash') }}</dd>
          </div>
          <div v-if="pet.healthBookPdfAttached">
            <dt>{{ $t('clients.pet.healthBookPdf') }}</dt>
            <dd>
              <a
                class="pro-link"
                :href="`/api/pets/${petId}/health-book`"
                target="_blank"
                rel="noopener noreferrer"
                data-testid="pet-health-book-pdf"
              >{{ $t('clients.pet.healthBookOpenPdf') }}</a>
            </dd>
          </div>
          <div>
            <dt>{{ $t('clients.pet.summaryPlan') }}</dt>
            <dd>{{ petPlanLabel }}</dd>
          </div>
          <div>
            <dt>{{ $t('clients.pet.summaryStatus') }}</dt>
            <dd>
              <ProBadge :variant="petStatusVariant">{{ petStatusLabel }}</ProBadge>
            </dd>
          </div>
        </dl>
      </ProCard>
      <ProCard
        v-if="pet"
        :title="$t('clients.pet.lifecycleTitle')"
        class="pro-mb-lg"
        data-testid="pet-lifecycle"
      >
        <div class="pro-pet-horse-reg">
          <div>
            <label class="pro-label" for="pet-adopted-at">{{ $t('clients.pet.adoptedAt') }}</label>
            <input
              id="pet-adopted-at"
              v-model="lifecycleAdoptedAt"
              class="pro-input"
              type="date"
              data-testid="pet-adopted-at"
              :disabled="!canWriteClinical || lifecycleSaving"
            >
          </div>
          <div>
            <label class="pro-label" for="pet-sold-at">{{ $t('clients.pet.soldAt') }}</label>
            <input
              id="pet-sold-at"
              v-model="lifecycleSoldAt"
              class="pro-input"
              type="date"
              data-testid="pet-sold-at"
              :disabled="!canWriteClinical || lifecycleSaving"
            >
          </div>
          <div>
            <label class="pro-label" for="pet-deceased-at">{{ $t('clients.pet.deceasedAt') }}</label>
            <input
              id="pet-deceased-at"
              v-model="lifecycleDeceasedAt"
              class="pro-input"
              type="date"
              data-testid="pet-deceased-at"
              :disabled="!canWriteClinical || lifecycleSaving"
            >
          </div>
        </div>
        <p v-if="lifecycleError" class="pro-alert" data-testid="pet-lifecycle-error">{{ lifecycleError }}</p>
        <p v-if="lifecycleSaved" class="pro-hint" data-testid="pet-lifecycle-saved">{{ $t('clients.pet.lifecycleSaved') }}</p>
        <ProButton
          v-if="canWriteClinical"
          class="pro-mt-md"
          test-id="pet-lifecycle-save"
          :disabled="lifecycleSaving || !lifecycleDirty"
          @click="saveLifecycle"
        >
          {{ $t('common.save') }}
        </ProButton>
      </ProCard>
      <ProCard
        v-if="pharmacyEnabled && canReadPharmacy"
        :title="$t('clients.pet.dafDispensesTitle')"
        class="pro-mb-lg"
        data-testid="pet-daf-dispenses"
      >
        <p v-if="dafDispensesLoading" class="pro-hint">{{ $t('common.loading') }}</p>
        <p v-else-if="dafDispensesError" class="pro-error" role="alert">{{ dafDispensesError }}</p>
        <div v-else-if="!dafDispenses.length" class="pro-empty">{{ $t('clients.pet.dafDispensesEmpty') }}</div>
        <ul v-else class="pet-daf-dispenses">
          <li v-for="disp in dafDispenses" :key="disp.dafId">
            <NuxtLink :to="`/daf/${disp.dafId}`" class="pro-link">
              {{ disp.displayNumber || disp.dafId }}
            </NuxtLink>
            <span v-if="disp.finalizedAt" class="pro-hint"> · {{ formatDate(disp.finalizedAt) }}</span>
            <ul v-if="disp.items?.length" class="pet-daf-dispenses__items">
              <li v-for="(it, i) in disp.items" :key="i">
                {{ it.medicationName }} · {{ it.qty }} · lot {{ it.lotNumber || '—' }}
                <span v-if="it.ammNumber"> · AMM {{ it.ammNumber }}</span>
              </li>
            </ul>
          </li>
        </ul>
      </ProCard>
      <ProCard
        v-if="pet && isFoodChainSpecies(pet.species)"
        :title="$t('clients.pet.horseRegulatoryTitle')"
        class="pro-mb-lg"
        data-testid="pet-horse-regulatory"
      >
        <p class="pro-hint pro-mb-md">{{ $t('clients.pet.horseRegulatoryHint') }}</p>
        <div class="pro-pet-horse-reg">
          <div>
            <label class="pro-label" for="pet-food-chain">{{ $t('clients.pet.foodChainStatus') }}</label>
            <select
              id="pet-food-chain"
              v-model="horseFoodChainYesNo"
              class="pro-input"
              data-testid="pet-food-chain"
              :disabled="!canWriteClinical || horseRegSaving"
            >
              <option value="yes">{{ $t('clients.pet.foodChainYes') }}</option>
              <option value="no">{{ $t('clients.pet.foodChainNo') }}</option>
            </select>
          </div>
          <div>
            <label class="pro-label" for="pet-domicile">{{ $t('clients.pet.domicileLocation') }}</label>
            <input
              id="pet-domicile"
              v-model="horseDomicile"
              class="pro-input"
              type="text"
              maxlength="500"
              :placeholder="$t('clients.pet.domicilePlaceholder')"
              data-testid="pet-domicile"
              :disabled="!canWriteClinical || horseRegSaving"
            >
          </div>
        </div>
        <p v-if="horseRegError" class="pro-alert" data-testid="pet-horse-reg-error">{{ horseRegError }}</p>
        <p v-if="horseRegSaved" class="pro-hint" data-testid="pet-horse-reg-saved">{{ $t('clients.pet.horseRegulatorySaved') }}</p>
        <ProButton
          v-if="canWriteClinical"
          class="pro-mt-md"
          test-id="pet-horse-reg-save"
          :disabled="horseRegSaving || !horseRegDirty"
          @click="saveHorseRegulatory"
        >
          {{ $t('common.save') }}
        </ProButton>
      </ProCard>
      <ProCard
        v-if="hasAnyVitalsData"
        :title="$t('clients.pet.chartsToggle')"
        class="pro-mb-lg"
        data-testid="pet-overview-charts"
      >
        <p class="pro-hint pro-mb-md">{{ $t('clients.pet.chartsPreviewHint') }}</p>
        <ul class="pro-pet-charts-preview" :aria-label="$t('clients.pet.chartsToggle')">
          <li v-if="hasChartData">
            <ProIcon name="favorite" :size="18" aria-hidden="true" />
            <span>{{ $t('clients.pet.chartsTypeHeartrate') }}</span>
            <strong>{{ kpiLastBpm }}</strong>
          </li>
          <li v-if="hasWeightChartData">
            <ProIcon name="monitor_weight" :size="18" aria-hidden="true" />
            <span>{{ $t('clients.pet.chartsTypeWeight') }}</span>
            <strong>{{ kpiLastWeight }}</strong>
          </li>
          <li v-if="hasBpChartData">
            <ProIcon name="monitor_heart" :size="18" aria-hidden="true" />
            <span>{{ $t('clients.pet.chartsTypeBp') }}</span>
            <strong>{{ kpiLastBp }}</strong>
          </li>
          <li v-if="hasLabChartData">
            <ProIcon name="science" :size="18" aria-hidden="true" />
            <span>{{ $t('clients.pet.chartsTypeLabs') }}</span>
            <strong>{{ kpiLastLab }}</strong>
          </li>
        </ul>
        <ProButton
          variant="secondary"
          test-id="pet-overview-open-vitals"
          @click="openVitalsTab"
        >
          {{ $t('clients.pet.chartsOpenVitals') }}
        </ProButton>
      </ProCard>

      <ProPetDayTimeline
        :items="timeline"
        class="pro-mb-lg"
        @open-visit="openVisitReport"
      />
    </div>

    <div
      v-show="activeTab === 'vitals'"
      role="tabpanel"
      aria-labelledby="tab-vitals"
      data-testid="pet-tab-vitals"
    >
      <ProPetVitalsCharts
        show-empty
        :chart-range="chartRange"
        :weight-chart-range="weightChartRange"
        :bp-chart-range="bpChartRange"
        :lab-chart-range="labChartRange"
        :has-chart-data="hasChartData"
        :has-weight-chart-data="hasWeightChartData"
        :has-bp-chart-data="hasBpChartData"
        :has-lab-chart-data="hasLabChartData"
        :chart-values="chartValues"
        :chart-alerts="chartAlerts"
        :chart-dates="chartDates"
        :domain-start="chartDomain.start"
        :domain-end="chartDomain.end"
        :weight-chart-values="weightChartValues"
        :weight-chart-dates="weightChartDates"
        :weight-domain-start="weightChartDomain.start"
        :weight-domain-end="weightChartDomain.end"
        :bp-chart-sys-values="bpChartSysValues"
        :bp-chart-dia-values="bpChartDiaValues"
        :bp-chart-dates="bpChartDates"
        :bp-domain-start="bpChartDomain.start"
        :bp-domain-end="bpChartDomain.end"
        :lab-trend-analyte="labTrendAnalyte"
        :lab-analyte-options="labAnalyteOptions"
        :lab-chart-values="labChartValues"
        :lab-chart-dates="labChartDates"
        :lab-domain-start="labChartDomain.start"
        :lab-domain-end="labChartDomain.end"
        :lab-axis-title="labTrendAnalyte ? labAnalyteLabel(labTrendAnalyte) : ''"
        @update:chart-range="chartRange = $event"
        @update:weight-chart-range="weightChartRange = $event"
        @update:bp-chart-range="bpChartRange = $event"
        @update:lab-chart-range="labChartRange = $event"
        @update:lab-trend-analyte="onLabTrendAnalyte"
      />

      <ProCard :title="$t('clients.pet.heartrateTitle')" class="pro-mb-lg">
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.heartrateTitle')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': sessionFilter === 'all' }"
          data-testid="pet-filter-all"
          @click="sessionFilter = 'all'"
        >
          {{ $t('clients.pet.filterAll') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': sessionFilter === 'alerts' }"
          data-testid="pet-filter-alerts"
          @click="sessionFilter = 'alerts'"
        >
          {{ $t('clients.pet.filterAlerts') }}
        </button>
      </div>
      <ProTable
        :empty="!filteredSessions.length"
        :empty-title="$t('clients.pet.heartrateEmptyTitle')"
        :empty-description="$t('clients.pet.heartrateEmptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.columnBpm') }}</th>
            <th>{{ $t('clients.pet.columnDuration') }}</th>
            <th>{{ $t('clients.pet.columnStatus') }}</th>
            <th>{{ $t('clients.pet.columnComment') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in filteredSessions"
            :key="s.id"
            :class="{ 'pro-table-row--alert': s.isAlert }"
          >
            <td>
              <span class="pro-pet-reading-date">
                {{ formatDate(s.startedAt) }}
                <ProBadge
                  v-if="isReadingNew(s)"
                  variant="danger"
                  data-testid="reading-new-badge"
                >
                  {{ $t('clients.pet.newReading') }}
                </ProBadge>
              </span>
            </td>
            <td><code>{{ s.bpm }}</code></td>
            <td>{{ s.durationSec }}s</td>
            <td>
              <ProBadge :variant="s.isAlert ? 'danger' : 'success'">
                {{ s.isAlert ? $t('clients.pet.alert') : $t('clients.pet.ok') }}
              </ProBadge>
            </td>
            <td>
              <span
                v-if="s.comment"
                class="pro-pet-reading-comment"
                data-testid="pet-reading-comment"
                :title="s.comment"
              >{{ s.comment }}</span>
              <span v-else class="text-muted">—</span>
            </td>
          </tr>
        </tbody>
      </ProTable>
      </ProCard>

      <ProCard :title="$t('clients.pet.weightTitle')" data-testid="pet-weight-table-card" class="pro-mb-lg">
      <ProTable
        :empty="!weights.length"
        :empty-title="$t('clients.pet.weightEmptyTitle')"
        :empty-description="$t('clients.pet.weightEmptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.columnWeight') }}</th>
            <th>{{ $t('clients.pet.columnComment') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="w in weights" :key="w.id">
            <td>{{ formatDate(w.recordedAt) }}</td>
            <td><code>{{ w.weightKg }} kg</code></td>
            <td>
              <span
                v-if="w.comment"
                class="pro-pet-reading-comment"
                data-testid="pet-weight-comment"
                :title="w.comment"
              >{{ w.comment }}</span>
              <span v-else class="text-muted">—</span>
            </td>
          </tr>
        </tbody>
      </ProTable>
      </ProCard>

      <ProCard :title="$t('clients.pet.bpTitle')" data-testid="pet-bp-table-card" class="pro-mb-lg">
        <form
          v-if="canWriteClinical"
          class="pro-pet-inline-form"
          data-testid="pet-bp-form"
          @submit.prevent="createBloodPressure"
        >
          <input
            id="pet-bp-sys"
            v-model.number="bpDraft.systolicMmHg"
            class="pro-input"
            type="number"
            min="20"
            max="400"
            required
            :aria-label="$t('clients.pet.bpSystolic')"
            :placeholder="$t('clients.pet.bpSystolic')"
            data-testid="pet-bp-sys"
          />
          <input
            id="pet-bp-dia"
            v-model.number="bpDraft.diastolicMmHg"
            class="pro-input"
            type="number"
            min="10"
            max="300"
            required
            :aria-label="$t('clients.pet.bpDiastolic')"
            :placeholder="$t('clients.pet.bpDiastolic')"
            data-testid="pet-bp-dia"
          />
          <select v-model="bpDraft.method" class="pro-input" :aria-label="$t('clients.pet.bpMethod')">
            <option value="doppler">{{ $t('clients.pet.bpMethodDoppler') }}</option>
            <option value="oscillometric">{{ $t('clients.pet.bpMethodOscillometric') }}</option>
            <option value="invasive">{{ $t('clients.pet.bpMethodInvasive') }}</option>
            <option value="unknown">{{ $t('clients.pet.bpMethodUnknown') }}</option>
          </select>
          <input
            v-model="bpDraft.site"
            class="pro-input"
            :aria-label="$t('clients.pet.bpSite')"
            :placeholder="$t('clients.pet.bpSite')"
            maxlength="80"
          />
          <input
            v-model="bpDraft.comment"
            class="pro-input"
            :aria-label="$t('clients.pet.bpComment')"
            :placeholder="$t('clients.pet.bpComment')"
            data-testid="pet-bp-comment"
            maxlength="500"
          />
          <ProButton type="submit" :disabled="bpBusy" test-id="pet-bp-save">
            {{ $t('clients.pet.bpAdd') }}
          </ProButton>
        </form>
        <p v-if="bpError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ bpError }}</p>
        <ProTable
          :empty="!bloodPressures.length"
          :empty-title="$t('clients.pet.bpEmptyTitle')"
          :empty-description="$t('clients.pet.bpEmptyDescription')"
        >
          <thead>
            <tr>
              <th>{{ $t('clients.pet.columnDate') }}</th>
              <th>{{ $t('clients.pet.bpTitle') }}</th>
              <th>{{ $t('clients.pet.bpMethod') }}</th>
              <th>{{ $t('clients.pet.bpSite') }}</th>
              <th>{{ $t('clients.pet.columnComment') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="bp in bloodPressures" :key="bp.id">
              <td>{{ formatDate(bp.recordedAt) }}</td>
              <td><code>{{ bp.systolicMmHg }}/{{ bp.diastolicMmHg }}</code></td>
              <td>{{ bpMethodLabel(bp.method) }}</td>
              <td>{{ bp.site || '—' }}</td>
              <td>
                <span v-if="bp.comment" class="pro-pet-reading-comment" :title="bp.comment">{{ bp.comment }}</span>
                <span v-else class="text-muted">—</span>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <ProCard :title="$t('clients.pet.labsTitle')" data-testid="pet-labs-card">
        <form
          v-if="canWriteClinical"
          class="pro-pet-labs-form"
          data-testid="pet-labs-form"
          @submit.prevent="createLabPanel"
        >
          <div class="pro-pet-inline-form">
            <input
              v-model="labDraft.labName"
              class="pro-input"
              :aria-label="$t('clients.pet.labsLabName')"
              :placeholder="$t('clients.pet.labsLabName')"
              data-testid="pet-labs-name"
            />
            <input
              v-model="labDraft.collectedAt"
              class="pro-input"
              type="datetime-local"
              :aria-label="$t('clients.pet.labsCollectedAt')"
            />
            <input
              v-model="labDraft.notes"
              class="pro-input"
              :aria-label="$t('clients.pet.labsNotes')"
              :placeholder="$t('clients.pet.labsNotes')"
            />
            <select
              v-model="labDraft.documentId"
              class="pro-input"
              data-testid="pet-labs-document"
              :aria-label="$t('clients.pet.labsDocument')"
            >
              <option value="">{{ $t('clients.pet.labsDocumentNone') }}</option>
              <option v-for="d in documents" :key="d.id" :value="d.id">
                {{ d.title || d.fileName }}
              </option>
            </select>
          </div>
          <div
            v-for="(row, idx) in labDraft.results"
            :key="idx"
            class="pro-pet-inline-form"
          >
            <select
              v-model="row.analyteCode"
              class="pro-input"
              required
              :aria-label="$t('clients.pet.labsAnalyte')"
            >
              <option disabled value="">{{ $t('clients.pet.labsAnalyte') }}</option>
              <option v-for="code in labAnalyteCodes" :key="code" :value="code">
                {{ labAnalyteLabel(code) }}
              </option>
            </select>
            <input
              v-model.number="row.valueNum"
              class="pro-input"
              type="number"
              step="any"
              :aria-label="$t('clients.pet.labsValue')"
              :placeholder="$t('clients.pet.labsValue')"
            />
            <input
              v-model="row.valueText"
              class="pro-input"
              :aria-label="$t('clients.pet.labsValueText')"
              :placeholder="$t('clients.pet.labsValueText')"
              maxlength="64"
            />
            <input
              v-model="row.unit"
              class="pro-input"
              :aria-label="$t('clients.pet.labsUnit')"
              :placeholder="$t('clients.pet.labsUnit')"
            />
            <input
              v-model.number="row.refLow"
              class="pro-input"
              type="number"
              step="any"
              :aria-label="$t('clients.pet.labsRef')"
              :placeholder="$t('clients.pet.labsRef') + ' ↓'"
            />
            <input
              v-model.number="row.refHigh"
              class="pro-input"
              type="number"
              step="any"
              :aria-label="$t('clients.pet.labsRef')"
              :placeholder="$t('clients.pet.labsRef') + ' ↑'"
            />
            <ProButton
              v-if="labDraft.results.length > 1"
              type="button"
              variant="ghost"
              :aria-label="$t('clients.pet.labsRemoveRow')"
              @click="labDraft.results.splice(idx, 1)"
            >
              ×
            </ProButton>
          </div>
          <div class="pro-pet-inline-form">
            <ProButton type="button" variant="secondary" @click="addLabResultRow">
              + {{ $t('clients.pet.labsAddRow') }}
            </ProButton>
            <ProButton type="submit" :disabled="labBusy" test-id="pet-labs-save">
              {{ $t('clients.pet.labsAdd') }}
            </ProButton>
          </div>
        </form>
        <p v-if="labError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ labError }}</p>

        <ProTable
          :empty="!labPanels.length"
          :empty-title="$t('clients.pet.labsEmptyTitle')"
          :empty-description="$t('clients.pet.labsEmptyDescription')"
        >
          <thead>
            <tr>
              <th>{{ $t('clients.pet.columnDate') }}</th>
              <th>{{ $t('clients.pet.labsLabName') }}</th>
              <th>{{ $t('clients.pet.labsFlag') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="panel in labPanels"
              :key="panel.id"
              class="pro-table-row--clickable"
              data-testid="pet-lab-panel-row"
              role="button"
              tabindex="0"
              :aria-label="$t('clients.pet.labsResults')"
              @click="openLabPanel(panel.id)"
              @keydown.enter.prevent="openLabPanel(panel.id)"
              @keydown.space.prevent="openLabPanel(panel.id)"
            >
              <td>{{ formatDate(panel.collectedAt) }}</td>
              <td>{{ panel.labName || '—' }}</td>
              <td>
                <ProBadge v-if="panel.abnormalCount > 0" variant="danger">
                  {{ $t('clients.pet.labsAbnormal', { n: panel.abnormalCount }) }}
                </ProBadge>
                <span v-else class="text-muted">—</span>
              </td>
              <td>
                <ProButton
                  v-if="canWriteClinical"
                  type="button"
                  variant="ghost"
                  test-id="pet-lab-delete"
                  @click.stop="deleteLabPanel(panel.id)"
                >
                  {{ $t('clients.pet.labsDelete') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>

        <div v-if="labDetail" class="pro-pet-lab-detail" data-testid="pet-lab-detail">
          <div class="pro-pet-inline-form pro-mb-md">
            <h3>{{ $t('clients.pet.labsResults') }} — {{ formatDate(labDetail.collectedAt) }}</h3>
            <ProButton
              v-if="canWriteClinical && !labEditing"
              type="button"
              variant="secondary"
              test-id="pet-lab-edit"
              @click="startLabEdit"
            >
              {{ $t('clients.pet.labsEdit') }}
            </ProButton>
          </div>
          <p v-if="labDetail.notes" class="text-muted">{{ labDetail.notes }}</p>
          <p v-if="labDetailDocument" class="pro-mb-md">
            <a
              :href="labDetailDocument.fileUrl"
              target="_blank"
              rel="noopener noreferrer"
              data-testid="pet-lab-document-link"
            >
              {{ $t('clients.pet.labsDocument') }} — {{ labDetailDocument.title || labDetailDocument.fileName }}
            </a>
          </p>

          <form
            v-if="labEditing"
            class="pro-pet-labs-form"
            data-testid="pet-lab-edit-form"
            @submit.prevent="saveLabEdit"
          >
            <div
              v-for="(row, idx) in labEditDraft.results"
              :key="idx"
              class="pro-pet-inline-form"
            >
              <select
                v-model="row.analyteCode"
                class="pro-input"
                required
                :aria-label="$t('clients.pet.labsAnalyte')"
              >
                <option disabled value="">{{ $t('clients.pet.labsAnalyte') }}</option>
                <option v-for="code in labAnalyteCodes" :key="code" :value="code">
                  {{ labAnalyteLabel(code) }}
                </option>
              </select>
              <input
                v-model.number="row.valueNum"
                class="pro-input"
                type="number"
                step="any"
                :aria-label="$t('clients.pet.labsValue')"
                :placeholder="$t('clients.pet.labsValue')"
                data-testid="pet-lab-edit-value"
              />
              <input
                v-model="row.valueText"
                class="pro-input"
                :aria-label="$t('clients.pet.labsValueText')"
                :placeholder="$t('clients.pet.labsValueText')"
                maxlength="64"
                data-testid="pet-lab-edit-value-text"
              />
              <input
                v-model="row.unit"
                class="pro-input"
                :aria-label="$t('clients.pet.labsUnit')"
                :placeholder="$t('clients.pet.labsUnit')"
              />
              <input
                v-model.number="row.refLow"
                class="pro-input"
                type="number"
                step="any"
                :aria-label="$t('clients.pet.labsRef')"
                :placeholder="$t('clients.pet.labsRef') + ' ↓'"
              />
              <input
                v-model.number="row.refHigh"
                class="pro-input"
                type="number"
                step="any"
                :aria-label="$t('clients.pet.labsRef')"
                :placeholder="$t('clients.pet.labsRef') + ' ↑'"
              />
              <ProButton
                v-if="labEditDraft.results.length > 1"
                type="button"
                variant="ghost"
                :aria-label="$t('clients.pet.labsRemoveRow')"
                @click="labEditDraft.results.splice(idx, 1)"
              >
                ×
              </ProButton>
            </div>
            <div class="pro-pet-inline-form">
              <ProButton type="button" variant="ghost" @click="addLabEditRow">
                + {{ $t('clients.pet.labsAddRow') }}
              </ProButton>
              <ProButton type="submit" :disabled="labBusy" test-id="pet-lab-save-results">
                {{ $t('clients.pet.labsSaveResults') }}
              </ProButton>
              <ProButton
                type="button"
                variant="secondary"
                :disabled="labBusy"
                test-id="pet-lab-cancel-edit"
                @click="cancelLabEdit"
              >
                {{ $t('clients.pet.labsCancelEdit') }}
              </ProButton>
            </div>
          </form>

          <ProTable v-else>
            <thead>
              <tr>
                <th>{{ $t('clients.pet.labsAnalyte') }}</th>
                <th>{{ $t('clients.pet.labsValue') }}</th>
                <th>{{ $t('clients.pet.labsUnit') }}</th>
                <th>{{ $t('clients.pet.labsRef') }}</th>
                <th>{{ $t('clients.pet.labsFlag') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in labDetail.results || []" :key="r.id">
                <td>{{ labAnalyteLabel(r.analyteCode) }}</td>
                <td><code>{{ r.valueNum ?? r.valueText ?? '—' }}</code></td>
                <td>{{ r.unit || '—' }}</td>
                <td>
                  <span v-if="r.refLow != null || r.refHigh != null">
                    {{ r.refLow ?? '…' }} – {{ r.refHigh ?? '…' }}
                  </span>
                  <span v-else>—</span>
                </td>
                <td>
                  <ProBadge :variant="labFlagVariant(r.flag)">{{ labFlagLabel(r.flag) }}</ProBadge>
                </td>
              </tr>
            </tbody>
          </ProTable>
        </div>
      </ProCard>
    </div>

    <div
      v-show="activeTab === 'care'"
      role="tabpanel"
      aria-labelledby="tab-care"
      data-testid="pet-tab-care"
    >
      <ProCard :title="$t('clients.pet.careTitle')" class="pro-mb-lg">
      <form v-if="canManageCare" class="pro-pet-inline-form" @submit.prevent="createCare">
        <input v-model="careDraft.title" class="pro-input" :placeholder="$t('clients.pet.careTitleField')" required />
        <select v-model="careDraft.type" class="pro-input" :aria-label="$t('clients.pet.careType')">
          <option value="vaccination">{{ $t('clients.pet.careTypeVaccination') }}</option>
          <option value="deworming">{{ $t('clients.pet.careTypeDeworming') }}</option>
          <option value="vet_check">{{ $t('clients.pet.careTypeVetCheck') }}</option>
          <option value="dental">{{ $t('clients.pet.careTypeDental') }}</option>
          <option value="farrier">{{ $t('clients.pet.careTypeFarrier') }}</option>
          <option value="fecal_egg">{{ $t('clients.pet.careTypeFecalEgg') }}</option>
          <option value="custom">{{ $t('clients.pet.careTypeCustom') }}</option>
        </select>
        <ProButton type="submit" :disabled="careBusy">{{ $t('clients.pet.careCreate') }}</ProButton>
      </form>
      <ProEmptyState
        v-if="!careReminders.length"
        :title="$t('clients.pet.careEmptyTitle')"
        :description="$t('clients.pet.careEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.careType') }}</th>
            <th>{{ $t('clients.pet.careTitleField') }}</th>
            <th>{{ $t('clients.pet.careDue') }}</th>
            <th>{{ $t('clients.pet.columnStatus') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in careReminders" :key="c.id">
            <td>{{ careTypeLabel(c.type) }}</td>
            <td>{{ c.title }}</td>
            <td>
              {{ formatDate(c.dueAt) }}
              <ProBadge v-if="isCareOverdue(c)" variant="danger">{{ $t('clients.pet.careOverdue') }}</ProBadge>
            </td>
            <td>{{ c.status }}</td>
            <td>
              <div v-if="canManageCare && c.status === 'pending'" class="pro-flex-gap">
                <ProIconAction
                  icon="task_alt"
                  :label="$t('clients.pet.careDone')"
                  :disabled="careBusy"
                  @click="markCareDone(c.id)"
                />
                <ProIconAction
                  icon="schedule"
                  :label="$t('clients.pet.carePostpone')"
                  :disabled="careBusy"
                  @click="postponeCare(c.id, 7)"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
      </ProCard>

      <ProCard :title="$t('clients.pet.visitsTitle')" class="pro-mb-lg">
      <form v-if="canManageCalendar" class="pro-pet-inline-form" @submit.prevent="proposeVisit(false)">
        <input v-model="visitDraft.scheduledAt" class="pro-input" type="datetime-local" :aria-label="$t('clients.pet.visitScheduledAt')" required />
        <select v-model="visitDraft.visitTypeId" class="pro-select" :aria-label="$t('calendar.visitType')" data-testid="pet-visit-type">
          <option value="">{{ $t('calendar.visitTypeNone') }}</option>
          <option v-for="vt in visitTypes" :key="vt.id" :value="vt.id">
            {{ vt.name }} ({{ vt.durationMinutes }} min)
          </option>
        </select>
        <input
          v-model.number="visitDraft.durationMinutes"
          class="pro-input"
          type="number"
          min="5"
          max="480"
          step="5"
          :disabled="!!visitDraft.visitTypeId"
          :aria-label="$t('calendar.durationMinutes')"
          data-testid="pet-visit-duration"
        >
        <input v-model="visitDraft.notes" class="pro-input" :placeholder="$t('clients.pet.visitNotes')" />
        <label class="pro-checkbox-label" data-testid="visit-request-preconsult">
          <input v-model="visitDraft.requestPreconsult" type="checkbox" class="pro-checkbox">
          {{ $t('clients.pet.visitRequestPreconsult') }}
        </label>
        <ProButton type="submit" :disabled="visitBusy" variant="secondary">
          {{ $t('clients.pet.visitPropose') }}
        </ProButton>
        <ProButton type="button" :disabled="visitBusy || !visitDraft.scheduledAt" @click="proposeVisit(true)">
          {{ $t('clients.pet.visitConfirmDirect') }}
        </ProButton>
      </form>
      <ProEmptyState
        v-if="!visits.length"
        :title="$t('clients.pet.visitsEmptyTitle')"
        :description="$t('clients.pet.visitsEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.visitNotes') }}</th>
            <th>{{ $t('clients.pet.visitStatus') }}</th>
            <th>{{ $t('clients.pet.visitSource') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in visits" :key="v.id">
            <td>{{ formatDate(v.scheduledAt || v.createdAt) }}</td>
            <td>{{ v.notes || '—' }}</td>
            <td>{{ v.status }}</td>
            <td>
              {{ v.source === 'vet' ? $t('clients.pet.visitSourceVet') : $t('clients.pet.visitSourceClient') }}
            </td>
            <td>
              <div class="pro-flex-gap">
                <template v-if="canManageCalendar && v.status === 'requested' && v.pendingActionBy === 'vet'">
                  <label class="pro-checkbox-label" data-testid="visit-confirm-request-preconsult">
                    <input
                      v-model="confirmPreconsultByVisit[v.id]"
                      type="checkbox"
                      class="pro-checkbox"
                    >
                    {{ $t('clients.pet.visitRequestPreconsult') }}
                  </label>
                  <ProIconAction
                    icon="check"
                    :label="$t('clients.pet.visitConfirm')"
                    :disabled="visitBusy"
                    @click="visitAction(v.id, 'confirm')"
                  />
                </template>
                <ProIconAction
                  v-if="canManageCalendar && v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  icon="event_available"
                  :label="$t('calendar.acceptReschedule')"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'accept_reschedule')"
                />
                <ProIconAction
                  v-if="canManageCalendar && v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  icon="close"
                  :label="$t('calendar.rejectReschedule')"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'reject_reschedule')"
                />
                <ProIconAction
                  icon="description"
                  :label="$t('calendar.reportTitle')"
                  :test-id="`pet-visit-report-open-${v.id}`"
                  @click="openVisitReport(v)"
                />
                <ProIconAction
                  v-if="canManageCalendar && v.status === 'confirmed'"
                  icon="task_alt"
                  :label="$t('clients.pet.visitDone')"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'done')"
                />
                <ProIconAction
                  v-if="canManageCalendar && (v.status === 'requested' || v.status === 'confirmed' || v.status === 'reschedule_pending')"
                  icon="cancel"
                  variant="danger"
                  :label="$t('clients.pet.visitCancel')"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'cancel')"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
      </ProCard>
    </div>

    <div
      v-show="activeTab === 'documents'"
      role="tabpanel"
      aria-labelledby="tab-documents"
      data-testid="pet-tab-documents"
    >
      <ProCard :title="$t('clients.pet.documentsTitle')" class="pro-mb-lg">
      <form v-if="canWriteClinical" class="pro-pet-inline-form" @submit.prevent="uploadDocument">
        <input
          ref="docInputEl"
          type="file"
          accept="application/pdf,image/jpeg,image/png,image/webp"
          class="pro-input"
          data-testid="pet-document-input"
          @change="onDocFile"
        />
        <input
          v-model="docTitle"
          class="pro-input"
          :placeholder="$t('clients.pet.documentTitlePlaceholder')"
          data-testid="pet-document-title"
        />
        <ProButton type="submit" :disabled="docBusy || !docFile" data-testid="pet-document-upload">
          {{ $t('clients.pet.documentUpload') }}
        </ProButton>
      </form>
      <p v-if="docError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ docError }}</p>
      <p class="pro-hint">{{ $t('clients.pet.documentsHint') }}</p>
      <ProEmptyState
        v-if="!documents.length"
        :title="$t('clients.pet.documentsEmptyTitle')"
        :description="$t('clients.pet.documentsEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.documentName') }}</th>
            <th>{{ $t('clients.pet.documentType') }}</th>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.documentUploader') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in documents" :key="d.id">
            <td>
              <a :href="d.fileUrl" target="_blank" rel="noopener noreferrer">
                {{ d.title || d.fileName }}
              </a>
            </td>
            <td>{{ documentKindLabel(d.contentType) }}</td>
            <td>{{ formatDate(d.createdAt) }}</td>
            <td>{{ d.uploaderName || $t('common.dash') }}</td>
            <td>
              <div class="pro-flex-gap">
                <ProIconAction
                  icon="open_in_new"
                  :label="$t('clients.pet.documentOpen')"
                  :href="d.fileUrl"
                />
                <ProIconAction
                  v-if="canWriteClinical"
                  icon="delete"
                  variant="danger"
                  :label="$t('common.delete')"
                  :disabled="docBusy"
                  @click="deleteDocument(d.id)"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
      </ProCard>
    </div>

    <div
      v-if="pacsEnabled"
      v-show="activeTab === 'imaging'"
      role="tabpanel"
      aria-labelledby="tab-imaging"
      data-testid="pet-tab-imaging"
    >
      <ProCard class="pro-mb-lg">
        <PacsViewerContainer :pet-id="petId" />
      </ProCard>
    </div>

    <div
      v-show="activeTab === 'sharing'"
      role="tabpanel"
      aria-labelledby="tab-sharing"
      data-testid="pet-tab-sharing"
    >
      <ProCard :title="$t('share.petTitle')" class="pro-mb-lg" data-testid="pet-shares-card">
      <p class="pro-hint pro-mb-md">{{ $t('share.petHint') }}</p>
      <form v-if="canManageShares" class="pro-pet-inline-form" @submit.prevent="addPetShare">
        <select v-model="shareColleagueId" class="pro-input" data-testid="pet-share-colleague">
          <option value="">{{ $t('share.colleaguePlaceholder') }}</option>
          <option v-for="c in colleagues" :key="c.userId" :value="c.userId">
            {{ c.fullName }} ({{ c.email }})
          </option>
        </select>
        <ProInput v-model="shareEmail" type="email" :label="$t('share.email')" test-id="pet-share-email" />
        <select v-model="sharePermission" class="pro-input" data-testid="pet-share-permission">
          <option value="read">{{ $t('share.permRead') }}</option>
          <option value="write_notes">{{ $t('share.permWriteNotes') }}</option>
          <option value="full">{{ $t('share.permFull') }}</option>
        </select>
        <select v-model="shareExpiresDays" class="pro-input" data-testid="pet-share-expires">
          <option value="">{{ $t('share.expiresNever') }}</option>
          <option value="7">{{ $t('share.expiresDays', { n: 7 }) }}</option>
          <option value="30">{{ $t('share.expiresDays', { n: 30 }) }}</option>
          <option value="90">{{ $t('share.expiresDays', { n: 90 }) }}</option>
        </select>
        <ProButton type="submit" :disabled="shareBusy" test-id="pet-share-submit">
          {{ $t('share.add') }}
        </ProButton>
      </form>
      <p v-if="shareError" class="pro-error">{{ shareError }}</p>
      <ProTable v-if="petShares.length">
        <thead>
          <tr>
            <th>{{ $t('share.columnName') }}</th>
            <th>{{ $t('share.columnEmail') }}</th>
            <th>{{ $t('share.columnPermission') }}</th>
            <th>{{ $t('share.columnExpires') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in petShares" :key="s.id">
            <td>{{ s.granteeName }}</td>
            <td>{{ s.granteeEmail }}</td>
            <td>{{ s.permission }}</td>
            <td>{{ s.expiresAt ? formatShareDate(s.expiresAt) : $t('share.expiresNever') }}</td>
            <td>
              <ProButton
                v-if="canManageShares"
                variant="ghost"
                :disabled="shareBusy"
                @click="revokePetShare(s.granteeUserId)"
              >
                {{ $t('share.revoke') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
      </ProCard>
    </div>

    <ProModal
      v-model:open="visitReportOpen"
      :title="$t('calendar.reportTitle')"
      size="lg"
      contain-scroll
    >
      <div class="pet-visit-report-modal">
        <ProVisitReportPanel
          v-if="visitReportId"
          fill-height
          :visit-id="visitReportId"
          :visit-scheduled-at="visitReportScheduledAt"
          :readonly="!canWriteClinical"
        />
        <p class="pro-hint pet-visit-report-modal__link">
          <NuxtLink
            v-if="visitReportId"
            :to="`/calendar?visit=${visitReportId}`"
            data-testid="pet-visit-report-calendar-link"
          >
            {{ $t('clients.pet.openReportInCalendar') }}
          </NuxtLink>
        </p>
      </div>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import { isFoodChainSpecies } from '~/utils/pet-species'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pets.read' })

const config = useRuntimeConfig()
const pharmacyEnabled = computed(() => isPublicFlagOn(config.public.pharmacyEnabled))
const pacsEnabled = computed(() => isPublicFlagOn(config.public.pacsEnabled))

const route = useRoute()
const { t, te } = useI18n()
const { canPractice } = usePracticePerms()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const canManageCare = computed(() => canPractice('care.manage'))
const canManageCalendar = computed(() => canPractice('calendar.manage'))
const canManageShares = computed(() => canPractice('shares.manage'))
const canReadShares = computed(() => canPractice('shares.read'))
const canMessage = computed(() => canPractice('messaging'))
const canValidateHR = computed(() => canPractice('heartrate.validate'))
const canReadPharmacy = computed(() => canPractice('pharmacy.read'))
const activeConsult = useActiveConsultation()
const clientId = route.params.clientId as string
const petId = route.params.petId as string

function speciesLabel(species: string | null | undefined) {
  if (!species) return t('common.dash')
  const key = `common.species.${species}`
  return te(key) ? t(key) : species
}

function openConsultation() {
  if (!canWriteClinical.value) return
  activeConsult.openForClient(clientId, petId)
}
const pet = ref<any>(null)
const petPhotoUrl = ref('')
const horseFoodChainYesNo = ref<'yes' | 'no'>('no')
const horseDomicile = ref('')
const horseRegSaving = ref(false)
const horseRegError = ref('')
const horseRegSaved = ref(false)
const lifecycleAdoptedAt = ref('')
const lifecycleSoldAt = ref('')
const lifecycleDeceasedAt = ref('')
const lifecycleSaving = ref(false)
const lifecycleError = ref('')
const lifecycleSaved = ref(false)

function toDateInputValue(value?: string | null) {
  if (!value) return ''
  const day = String(value).slice(0, 10)
  return /^\d{4}-\d{2}-\d{2}$/.test(day) ? day : ''
}

function formatPetDay(value?: string | null) {
  if (!value) return t('common.dash')
  const day = String(value).slice(0, 10)
  if (/^\d{4}-\d{2}-\d{2}$/.test(day)) {
    const [y, m, d] = day.split('-')
    return `${d}/${m}/${y}`
  }
  return formatDate(value)
}

function foodChainToYesNo(status?: string | null): 'yes' | 'no' {
  return status === 'food_producing' ? 'yes' : 'no'
}

function yesNoToFoodChain(v: 'yes' | 'no') {
  return v === 'yes' ? 'food_producing' : 'companion'
}
const sessions = ref<any[]>([])
const weights = ref<any[]>([])
const bloodPressures = ref<any[]>([])
const labPanels = ref<any[]>([])
const labDetail = ref<any | null>(null)
const timeline = ref<any[]>([])
const careReminders = ref<any[]>([])
const visits = ref<any[]>([])
const documents = ref<any[]>([])
const sessionFilter = ref<'all' | 'alerts'>('all')
const chartRange = ref<'3m' | '6m' | '1y'>('3m')
const weightChartRange = ref<'3m' | '6m' | '1y'>('3m')
const bpChartRange = ref<'3m' | '6m' | '1y'>('3m')
const labChartRange = ref<'3m' | '6m' | '1y'>('1y')
const bpBusy = ref(false)
const bpError = ref('')
const bpDraft = reactive({
  systolicMmHg: 120,
  diastolicMmHg: 80,
  method: 'doppler',
  site: '',
  comment: '',
})
const labEditing = ref(false)
const labEditDraft = reactive({
  results: [] as Array<{
    analyteCode: string
    valueNum: number | null
    valueText: string
    unit: string
    refLow: number | null
    refHigh: number | null
  }>,
})
const labBusy = ref(false)
const labError = ref('')
const labTrendAnalyte = ref('')
const labTrendPoints = ref<any[]>([])
const labAnalyteCodes = [
  'crea', 'urea', 'bun', 'alat', 'alt', 'asat', 'ast', 'alp', 'ggt',
  'hct', 'wbc', 'plt', 'glucose', 'tp', 'protein',
] as const
const labDraft = reactive({
  labName: '',
  collectedAt: '',
  notes: '',
  documentId: '',
  results: [{
    analyteCode: 'crea',
    valueNum: null as number | null,
    valueText: '',
    unit: 'mg/dL',
    refLow: null as number | null,
    refHigh: null as number | null,
  }],
})

const labDetailDocument = computed(() => {
  const id = labDetail.value?.documentId
  if (!id) return null
  return documents.value.find((d: any) => d.id === id) || null
})
const highlightedNewIds = ref<Set<string>>(new Set())
const careBusy = ref(false)
const visitBusy = ref(false)
const visitReportOpen = ref(false)
const visitReportId = ref('')
const visitReportScheduledAt = ref('')
const messagingBusy = ref(false)
const messagingError = ref('')
const pageError = ref('')
const docBusy = ref(false)
const docError = ref('')
const docTitle = ref('')
const docFile = ref<File | null>(null)
const docInputEl = ref<HTMLInputElement | null>(null)
const petShares = ref<any[]>([])
const colleagues = ref<{ userId: string; fullName: string; email: string }[]>([])
const shareColleagueId = ref('')
const shareEmail = ref('')
const sharePermission = ref('write_notes')
const shareExpiresDays = ref('')
const shareBusy = ref(false)
const shareError = ref('')
const careDraft = reactive({ title: '', type: 'vaccination' })
const visitDraft = reactive({
  scheduledAt: '',
  notes: '',
  requestPreconsult: false,
  visitTypeId: '',
  durationMinutes: 30,
})
const visitTypes = ref<{ id: string; name: string; durationMinutes: number }[]>([])
const dafDispenses = ref<Array<{
  dafId: string
  displayNumber?: string
  finalizedAt?: string
  items?: Array<{ medicationName: string; lotNumber?: string; qty: number; ammNumber?: string }>
}>>([])
const dafDispensesLoading = ref(false)
const dafDispensesError = ref('')

watch(
  () => visitDraft.visitTypeId,
  (id) => {
    if (!id) return
    const vt = visitTypes.value.find((x) => x.id === id)
    if (vt) visitDraft.durationMinutes = vt.durationMinutes
  },
)
const confirmPreconsultByVisit = reactive<Record<string, boolean>>({})
const activeTab = ref('overview')
let sessionsPollTimer: ReturnType<typeof setInterval> | null = null

const { formatDate } = useFormatters()
const { mapError } = useApiError()
const { user, fetchUser } = useProUser()
const { refresh: refreshNavBadges } = useNavBadges()
const router = useRouter()

const vitalsTabCount = computed(() => {
  const n = sessions.value.length
    + weights.value.length
    + bloodPressures.value.length
    + labPanels.value.length
  return n || undefined
})

const petTabs = computed(() => {
  const tabs = [
    { id: 'overview', label: t('clients.pet.tabs.overview') },
    { id: 'vitals', label: t('clients.pet.tabs.vitals'), count: vitalsTabCount.value },
    { id: 'care', label: t('clients.pet.tabs.care'), count: careReminders.value.length || undefined },
    { id: 'documents', label: t('clients.pet.tabs.documents'), count: documents.value.length || undefined },
  ]
  if (pacsEnabled.value) {
    tabs.push({ id: 'imaging', label: t('clients.pet.tabs.imaging') })
  }
  if (canReadShares.value) {
    tabs.push({ id: 'sharing', label: t('clients.pet.tabs.sharing') })
  }
  return tabs
})

const petPlanLabel = computed(() => {
  const code = pet.value?.entitlement?.planCode || pet.value?.planCode
  if (!code) return t('common.dash')
  const key = `products.plans.${code}.name`
  const translated = t(key)
  return translated !== key ? translated : String(code)
})

const petStatusLabel = computed(() => {
  const status = pet.value?.entitlement?.status || pet.value?.paymentStatus || ''
  if (!status) return t('common.dash')
  const key = `clients.pet.paymentStatus.${status}`
  const translated = t(key)
  return translated !== key ? translated : status
})

const petStatusVariant = computed(() => {
  const status = pet.value?.entitlement?.status || pet.value?.paymentStatus || ''
  if (status === 'active') return 'success'
  if (status === 'pending' || status === 'pending_payment') return 'warning'
  if (status === 'expired' || status === 'cancelled') return 'danger'
  return 'neutral'
})

const kpiLastBpm = computed(() => {
  const s = sessions.value[0]
  return s?.bpm != null ? String(s.bpm) : '—'
})

const kpiAlertCount = computed(() => {
  const cutoff = Date.now() - 7 * 24 * 60 * 60 * 1000
  return sessions.value.filter(s =>
    s.isAlert && s.startedAt && +new Date(s.startedAt) >= cutoff,
  ).length
})

const sortedWeights = computed(() =>
  [...weights.value].sort((a, b) => +new Date(b.recordedAt) - +new Date(a.recordedAt)),
)

const kpiLastWeight = computed(() => {
  const w = sortedWeights.value[0]
  return w?.weightKg != null ? `${w.weightKg} kg` : '—'
})

const kpiWeightDelta = computed(() => {
  const [latest, prev] = sortedWeights.value
  if (!latest || !prev || latest.weightKg == null || prev.weightKg == null) return ''
  const delta = Number(latest.weightKg) - Number(prev.weightKg)
  if (!Number.isFinite(delta) || delta === 0) return ''
  const sign = delta > 0 ? '+' : ''
  return `${sign}${delta.toFixed(1)} kg`
})

const sortedBloodPressures = computed(() =>
  [...bloodPressures.value].sort((a, b) => +new Date(b.recordedAt) - +new Date(a.recordedAt)),
)

const kpiLastBp = computed(() => {
  const bp = sortedBloodPressures.value[0]
  if (!bp) return '—'
  return `${bp.systolicMmHg}/${bp.diastolicMmHg}`
})

const kpiLastLab = computed(() => {
  const panel = labPanels.value[0]
  if (!panel?.collectedAt) return '—'
  return formatDate(panel.collectedAt)
})

const kpiOverdueCare = computed(() =>
  careReminders.value.filter(c => isCareOverdue(c)).length,
)

const kpiNextVisit = computed(() => {
  const upcoming = visits.value
    .filter(v =>
      v.scheduledAt
      && ['requested', 'confirmed', 'reschedule_pending'].includes(v.status)
      && +new Date(v.scheduledAt) >= Date.now(),
    )
    .sort((a, b) => +new Date(a.scheduledAt) - +new Date(b.scheduledAt))
  if (!upcoming.length) return '—'
  return formatDate(upcoming[0].scheduledAt)
})

function isReadingNew(s: { id: string, isNew?: boolean }) {
  return highlightedNewIds.value.has(s.id) || !!s.isNew
}

async function loadSessions(markSeenAfter = false) {
  const sessionsRes: any = await $fetch(`/api/pets/${petId}/heartrate`)
  const list = sessionsRes.data ?? sessionsRes ?? []
  const next = new Set(highlightedNewIds.value)
  let hadNew = false
  for (const s of list) {
    if (s.isNew) {
      next.add(s.id)
      hadNew = true
    }
  }
  highlightedNewIds.value = next
  sessions.value = list
  if (markSeenAfter && hadNew && canValidateHR.value) {
    try {
      await $fetch(`/api/pets/${petId}/heartrate/seen`, { method: 'POST' })
      await refreshNavBadges()
    } catch {
      // Non-blocking: unread badges may stay until next visit.
    }
  }
}

async function loadWeights() {
  const res: any = await $fetch(`/api/pets/${petId}/weights`)
  weights.value = res.data ?? res ?? []
}

async function loadBloodPressures() {
  const res: any = await $fetch(`/api/pets/${petId}/blood-pressure`)
  bloodPressures.value = res.data ?? res ?? []
}

async function preferLabTrendAnalyte() {
  if (labTrendAnalyte.value || !labPanels.value.length) return
  const latestId = labPanels.value[0]?.id as string | undefined
  if (!latestId) return
  try {
    const res: any = await $fetch(`/api/pets/${petId}/lab-panels/${latestId}`)
    const panel = res.data ?? res
    const results = Array.isArray(panel?.results) ? panel.results : []
    const withNum = results.find((r: any) =>
      r?.analyteCode && Number.isFinite(Number(r.valueNum)),
    )
    const code = String(withNum?.analyteCode || results[0]?.analyteCode || '').trim()
    if (!code) return
    labTrendAnalyte.value = code
    await loadLabTrend()
  } catch {
    // Leave analyte unset — user picks from the select.
  }
}

async function loadLabPanels() {
  const res: any = await $fetch(`/api/pets/${petId}/lab-panels`)
  labPanels.value = res.data ?? res ?? []
  await preferLabTrendAnalyte()
}

function openVitalsTab() {
  if (activeTab.value !== 'vitals') {
    activeTab.value = 'vitals'
  }
  router.replace({ query: { ...route.query, tab: 'vitals' } })
}

function bpMethodLabel(method: string) {
  switch (method) {
    case 'doppler': return t('clients.pet.bpMethodDoppler')
    case 'oscillometric': return t('clients.pet.bpMethodOscillometric')
    case 'invasive': return t('clients.pet.bpMethodInvasive')
    default: return t('clients.pet.bpMethodUnknown')
  }
}

function labAnalyteLabel(code: string) {
  const key = `clients.pet.labsAnalyte${code.charAt(0).toUpperCase()}${code.slice(1)}`
  const translated = t(key)
  return translated !== key ? translated : code.toUpperCase()
}

function labFlagLabel(flag: string) {
  switch (flag) {
    case 'low': return t('clients.pet.labsFlagLow')
    case 'high': return t('clients.pet.labsFlagHigh')
    case 'normal': return t('clients.pet.labsFlagNormal')
    default: return t('clients.pet.labsFlagUnknown')
  }
}

function labFlagVariant(flag: string): 'danger' | 'success' | 'neutral' {
  if (flag === 'low' || flag === 'high') return 'danger'
  if (flag === 'normal') return 'success'
  return 'neutral'
}

function addLabResultRow() {
  labDraft.results.push({
    analyteCode: '',
    valueNum: null,
    valueText: '',
    unit: '',
    refLow: null,
    refHigh: null,
  })
}

async function createBloodPressure() {
  if (!canWriteClinical.value) return
  const sys = Number(bpDraft.systolicMmHg)
  const dia = Number(bpDraft.diastolicMmHg)
  if (!Number.isFinite(sys) || !Number.isFinite(dia) || dia > sys || sys < 20 || dia < 10) {
    bpError.value = t('clients.pet.bpInvalid')
    return
  }
  bpBusy.value = true
  bpError.value = ''
  try {
    await $fetch(`/api/pets/${petId}/blood-pressure`, {
      method: 'POST',
      body: {
        systolicMmHg: sys,
        diastolicMmHg: dia,
        method: bpDraft.method,
        site: bpDraft.site.trim() || undefined,
        comment: bpDraft.comment.trim() || undefined,
      },
    })
    bpDraft.site = ''
    bpDraft.comment = ''
    await Promise.all([loadBloodPressures(), loadTimeline()])
  } catch (e: any) {
    bpError.value = mapError(e)
  } finally {
    bpBusy.value = false
  }
}

async function createLabPanel() {
  if (!canWriteClinical.value) return
  labBusy.value = true
  labError.value = ''
  try {
    const results: Array<Record<string, unknown>> = []
    for (const r of labDraft.results) {
      if (!r.analyteCode) continue
      const text = (r.valueText || '').trim()
      const hasNum = r.valueNum != null && Number.isFinite(Number(r.valueNum))
      if (!hasNum && !text) continue
      const row: Record<string, unknown> = {
        analyteCode: r.analyteCode,
        unit: r.unit || undefined,
        refLow: r.refLow != null && r.refLow !== ('' as any) ? Number(r.refLow) : undefined,
        refHigh: r.refHigh != null && r.refHigh !== ('' as any) ? Number(r.refHigh) : undefined,
      }
      if (hasNum) row.valueNum = Number(r.valueNum)
      if (text) row.valueText = text
      results.push(row)
    }
    if (!results.length) {
      labError.value = t('clients.pet.labsNeedResults')
      return
    }
    const body: Record<string, any> = {
      labName: labDraft.labName,
      notes: labDraft.notes,
      results,
    }
    if (labDraft.collectedAt) {
      body.collectedAt = new Date(labDraft.collectedAt).toISOString()
    }
    if (labDraft.documentId) {
      body.documentId = labDraft.documentId
    }
    await $fetch(`/api/pets/${petId}/lab-panels`, { method: 'POST', body })
    labDraft.labName = ''
    labDraft.notes = ''
    labDraft.collectedAt = ''
    labDraft.documentId = ''
    labDraft.results = [{
      analyteCode: 'crea',
      valueNum: null,
      valueText: '',
      unit: 'mg/dL',
      refLow: null,
      refHigh: null,
    }]
    labDetail.value = null
    await Promise.all([loadLabPanels(), loadTimeline()])
  } catch (e: any) {
    labError.value = mapError(e)
  } finally {
    labBusy.value = false
  }
}

async function openLabPanel(panelId: string) {
  if (labEditing.value) {
    if (!confirm(t('clients.pet.labsUnsavedConfirm'))) return
  }
  labEditing.value = false
  labEditDraft.results = []
  try {
    const res: any = await $fetch(`/api/pets/${petId}/lab-panels/${panelId}`)
    labDetail.value = res.data ?? res
  } catch (e: any) {
    labError.value = mapError(e)
  }
}

function emptyLabEditRow() {
  return {
    analyteCode: 'crea',
    valueNum: null as number | null,
    valueText: '',
    unit: 'mg/dL',
    refLow: null as number | null,
    refHigh: null as number | null,
  }
}

function startLabEdit() {
  if (!labDetail.value || !canWriteClinical.value) return
  const rows = (labDetail.value.results || []).map((r: any) => ({
    analyteCode: String(r.analyteCode || ''),
    valueNum: r.valueNum != null ? Number(r.valueNum) : null,
    valueText: String(r.valueText || ''),
    unit: String(r.unit || ''),
    refLow: r.refLow != null ? Number(r.refLow) : null,
    refHigh: r.refHigh != null ? Number(r.refHigh) : null,
  }))
  labEditDraft.results = rows.length ? rows : [emptyLabEditRow()]
  labEditing.value = true
}

function cancelLabEdit() {
  labEditing.value = false
  labEditDraft.results = []
}

function addLabEditRow() {
  labEditDraft.results.push(emptyLabEditRow())
}

async function saveLabEdit() {
  if (!canWriteClinical.value || !labDetail.value?.id) return
  const results: Array<Record<string, unknown>> = []
  for (const r of labEditDraft.results) {
    if (!r.analyteCode) continue
    const text = r.valueText.trim()
    const hasNum = r.valueNum != null && Number.isFinite(Number(r.valueNum))
    if (!hasNum && !text) continue
    const row: Record<string, unknown> = {
      analyteCode: r.analyteCode,
      unit: r.unit.trim() || undefined,
      refLow: r.refLow != null && Number.isFinite(Number(r.refLow)) ? Number(r.refLow) : undefined,
      refHigh: r.refHigh != null && Number.isFinite(Number(r.refHigh)) ? Number(r.refHigh) : undefined,
    }
    if (hasNum) row.valueNum = Number(r.valueNum)
    if (text) row.valueText = text
    results.push(row)
  }
  if (!results.length) {
    labError.value = t('clients.pet.labsNeedResults')
    return
  }
  labBusy.value = true
  labError.value = ''
  try {
    const res: any = await $fetch(`/api/pets/${petId}/lab-panels/${labDetail.value.id}`, {
      method: 'PATCH',
      body: { results },
    })
    labDetail.value = res.data ?? res
    labEditing.value = false
    labEditDraft.results = []
    await Promise.all([loadLabPanels(), loadTimeline()])
  } catch (e: any) {
    labError.value = mapError(e)
  } finally {
    labBusy.value = false
  }
}

async function deleteLabPanel(panelId: string) {
  if (!canWriteClinical.value) return
  if (!confirm(t('clients.pet.labsDeleteConfirm'))) return
  labBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/lab-panels/${panelId}`, { method: 'DELETE' })
    if (labDetail.value?.id === panelId) labDetail.value = null
    await Promise.all([loadLabPanels(), loadTimeline()])
  } catch (e: any) {
    labError.value = mapError(e)
  } finally {
    labBusy.value = false
  }
}

async function loadLabTrend() {
  if (!labTrendAnalyte.value) {
    labTrendPoints.value = []
    return
  }
  try {
    const res: any = await $fetch(`/api/pets/${petId}/lab-analytes/${labTrendAnalyte.value}/trend`)
    labTrendPoints.value = res.data ?? res ?? []
  } catch {
    labTrendPoints.value = []
  }
}

async function onLabTrendAnalyte(code: string) {
  labTrendAnalyte.value = code
  await loadLabTrend()
}

async function loadTimeline() {
  const res: any = await $fetch(`/api/pets/${petId}/timeline`)
  timeline.value = res.data ?? res ?? []
}

function careTypeLabel(type: string) {
  switch (type) {
    case 'vaccination':
      return t('clients.pet.careTypeVaccination')
    case 'deworming':
      return t('clients.pet.careTypeDeworming')
    case 'vet_check':
      return t('clients.pet.careTypeVetCheck')
    case 'dental':
      return t('clients.pet.careTypeDental')
    case 'farrier':
      return t('clients.pet.careTypeFarrier')
    case 'fecal_egg':
      return t('clients.pet.careTypeFecalEgg')
    case 'custom':
      return t('clients.pet.careTypeCustom')
    default:
      return type
  }
}

const petSubtitle = computed(() => {
  if (!pet.value) return ''
  return [speciesLabel(pet.value.species), pet.value.breed].filter(Boolean).join(' · ')
})

const isPrimaryPractice = computed(() => {
  const practiceId = user.value?.practiceId
  return Boolean(practiceId && pet.value?.practiceId && practiceId === pet.value.practiceId)
})

const filteredSessions = computed(() => {
  if (sessionFilter.value === 'alerts') {
    return sessions.value.filter((s) => s.isAlert)
  }
  return sessions.value
})

const recentAlert = computed(() => sessions.value.length > 0 && !!sessions.value[0]?.isAlert)

/** Approximate calendar months via 30-day units (avoids setMonth day overflow). */
function chartDomainFor(range: '3m' | '6m' | '1y') {
  const months = range === '3m' ? 3 : range === '6m' ? 6 : 12
  const end = new Date()
  const start = new Date(end.getTime() - months * 30 * 24 * 60 * 60 * 1000)
  return { start: start.toISOString(), end: end.toISOString() }
}

const chartDomain = computed(() => chartDomainFor(chartRange.value))

const hasChartData = computed(() =>
  sessions.value.some(s => s.bpm != null && s.startedAt),
)

const chartSessions = computed(() => {
  const cutoff = new Date(chartDomain.value.start)
  return [...sessions.value]
    .filter(s => s.bpm != null && s.startedAt && new Date(s.startedAt) >= cutoff)
    .sort((a, b) => +new Date(a.startedAt) - +new Date(b.startedAt))
})
const chartValues = computed(() => chartSessions.value.map(s => s.bpm as number))
const chartAlerts = computed(() => chartSessions.value.map(s => !!s.isAlert))
const chartDates = computed(() => chartSessions.value.map(s => s.startedAt as string))

const weightChartDomain = computed(() => chartDomainFor(weightChartRange.value))

const hasWeightChartData = computed(() =>
  weights.value.some(w => w.weightKg != null && w.recordedAt),
)

const weightChartReadings = computed(() => {
  const cutoff = new Date(weightChartDomain.value.start)
  return [...weights.value]
    .filter(w => w.weightKg != null && w.recordedAt && new Date(w.recordedAt) >= cutoff)
    .sort((a, b) => +new Date(a.recordedAt) - +new Date(b.recordedAt))
})
const weightChartValues = computed(() => weightChartReadings.value.map(w => Number(w.weightKg)))
const weightChartDates = computed(() => weightChartReadings.value.map(w => w.recordedAt as string))

const bpChartDomain = computed(() => chartDomainFor(bpChartRange.value))

const hasBpChartData = computed(() =>
  bloodPressures.value.some(bp => bp.systolicMmHg != null && bp.recordedAt),
)

const bpChartReadings = computed(() => {
  const cutoff = new Date(bpChartDomain.value.start)
  return [...bloodPressures.value]
    .filter(bp =>
      bp.systolicMmHg != null
      && bp.diastolicMmHg != null
      && bp.recordedAt
      && new Date(bp.recordedAt) >= cutoff,
    )
    .sort((a, b) => +new Date(a.recordedAt) - +new Date(b.recordedAt))
})
const bpChartSysValues = computed(() => bpChartReadings.value.map(bp => Number(bp.systolicMmHg)))
const bpChartDiaValues = computed(() => bpChartReadings.value.map(bp => Number(bp.diastolicMmHg)))
const bpChartDates = computed(() => bpChartReadings.value.map(bp => bp.recordedAt as string))

const hasLabChartData = computed(() => labPanels.value.length > 0)

const hasAnyVitalsData = computed(() =>
  hasChartData.value || hasWeightChartData.value || hasBpChartData.value || hasLabChartData.value,
)

const labAnalyteOptions = computed(() =>
  labAnalyteCodes.map(code => ({ code, label: labAnalyteLabel(code) })),
)

const labChartDomain = computed(() => chartDomainFor(labChartRange.value))

const labTrendSeries = computed(() => {
  const cutoff = new Date(labChartDomain.value.start)
  return labTrendPoints.value
    .filter(p =>
      Number.isFinite(Number(p.valueNum))
      && p.collectedAt
      && new Date(p.collectedAt) >= cutoff,
    )
    .sort((a, b) => +new Date(a.collectedAt) - +new Date(b.collectedAt))
})
const labChartValues = computed(() =>
  labTrendSeries.value.map(p => Number(p.valueNum)),
)
const labChartDates = computed(() =>
  labTrendSeries.value.map(p => p.collectedAt as string),
)

function onPetPhotoUploaded(data: any) {
  pet.value = { ...pet.value, ...data }
  petPhotoUrl.value = data?.photoUrl || petPhotoUrl.value
}

const horseRegDirty = computed(() => {
  if (!pet.value || !isFoodChainSpecies(pet.value.species)) return false
  const curYesNo = foodChainToYesNo(pet.value.foodChainStatus)
  const curDom = pet.value.domicileLocation || ''
  return horseFoodChainYesNo.value !== curYesNo || horseDomicile.value !== curDom
})

const lifecycleDirty = computed(() => {
  if (!pet.value) return false
  return (
    lifecycleAdoptedAt.value !== toDateInputValue(pet.value.adoptedAt)
    || lifecycleSoldAt.value !== toDateInputValue(pet.value.soldAt)
    || lifecycleDeceasedAt.value !== toDateInputValue(pet.value.deceasedAt)
  )
})

async function saveHorseRegulatory() {
  if (!canWriteClinical.value || !pet.value) return
  horseRegSaving.value = true
  horseRegError.value = ''
  horseRegSaved.value = false
  try {
    const status = yesNoToFoodChain(horseFoodChainYesNo.value)
    const res: any = await $fetch(`/api/vet/pets/${petId}/food-chain`, {
      method: 'PATCH',
      body: {
        foodChainStatus: status,
        domicileLocation: horseDomicile.value,
      },
    })
    const data = res?.data ?? res
    pet.value = {
      ...pet.value,
      foodChainStatus: data?.foodChainStatus ?? status,
      domicileLocation: data?.domicileLocation ?? horseDomicile.value,
    }
    horseFoodChainYesNo.value = foodChainToYesNo(pet.value.foodChainStatus)
    horseDomicile.value = pet.value.domicileLocation || ''
    horseRegSaved.value = true
  } catch (e: any) {
    horseRegError.value = mapError(e)
  } finally {
    horseRegSaving.value = false
  }
}

async function saveLifecycle() {
  if (!canWriteClinical.value || !pet.value) return
  lifecycleSaving.value = true
  lifecycleError.value = ''
  lifecycleSaved.value = false
  try {
    const res: any = await $fetch(`/api/vet/pets/${petId}/lifecycle`, {
      method: 'PATCH',
      body: {
        adoptedAt: lifecycleAdoptedAt.value,
        soldAt: lifecycleSoldAt.value,
        deceasedAt: lifecycleDeceasedAt.value,
      },
    })
    const data = res?.data ?? res
    pet.value = {
      ...pet.value,
      adoptedAt: data?.adoptedAt ?? (lifecycleAdoptedAt.value || null),
      soldAt: data?.soldAt ?? (lifecycleSoldAt.value || null),
      deceasedAt: data?.deceasedAt ?? (lifecycleDeceasedAt.value || null),
    }
    lifecycleAdoptedAt.value = toDateInputValue(pet.value.adoptedAt)
    lifecycleSoldAt.value = toDateInputValue(pet.value.soldAt)
    lifecycleDeceasedAt.value = toDateInputValue(pet.value.deceasedAt)
    lifecycleSaved.value = true
  } catch (e: any) {
    lifecycleError.value = mapError(e)
  } finally {
    lifecycleSaving.value = false
  }
}

function isCareOverdue(c: any) {
  return c.status === 'pending' && c.dueAt && new Date(c.dueAt).getTime() < Date.now()
}

async function loadCareAndVisits() {
  const [careRes, visitsRes]: any[] = await Promise.all([
    $fetch(`/api/pets/${petId}/care-reminders`),
    $fetch(`/api/pets/${petId}/visits`),
  ])
  careReminders.value = careRes.data ?? careRes ?? []
  visits.value = visitsRes.data ?? visitsRes ?? []
}

async function loadDocuments() {
  const res: any = await $fetch(`/api/pets/${petId}/documents`)
  documents.value = res.data ?? res ?? []
}

async function loadPetShares() {
  const [sharesRes, colleaguesRes]: any[] = await Promise.all([
    $fetch(`/api/pets/${petId}/shares`),
    $fetch('/api/vet/colleagues').catch(() => ({ data: [] })),
  ])
  petShares.value = sharesRes.data ?? sharesRes ?? []
  colleagues.value = colleaguesRes.data ?? colleaguesRes ?? []
}

async function addPetShare() {
  shareBusy.value = true
  shareError.value = ''
  try {
    const body: Record<string, string> = { permission: sharePermission.value || 'write_notes' }
    if (shareColleagueId.value) body.granteeUserId = shareColleagueId.value
    else if (shareEmail.value.trim()) body.email = shareEmail.value.trim()
    else {
      shareError.value = t('share.fieldsRequired')
      return
    }
    if (shareExpiresDays.value) {
      const d = new Date()
      d.setDate(d.getDate() + Number(shareExpiresDays.value))
      body.expiresAt = d.toISOString()
    }
    await $fetch(`/api/pets/${petId}/shares`, { method: 'POST', body })
    shareColleagueId.value = ''
    shareEmail.value = ''
    sharePermission.value = 'write_notes'
    shareExpiresDays.value = ''
    await loadPetShares()
  } catch (e: any) {
    shareError.value = mapError(e)
  } finally {
    shareBusy.value = false
  }
}

function formatShareDate(iso: string) {
  try {
    return new Date(iso).toLocaleDateString()
  } catch {
    return iso
  }
}

async function revokePetShare(granteeUserId: string) {
  shareBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/shares/${granteeUserId}`, { method: 'DELETE' })
    await loadPetShares()
  } catch (e: any) {
    shareError.value = mapError(e)
  } finally {
    shareBusy.value = false
  }
}

function documentKindLabel(contentType: string) {
  if (contentType?.includes('pdf')) return t('clients.pet.documentTypePdf')
  if (contentType?.startsWith('image/')) return t('clients.pet.documentTypeImage')
  return contentType || t('common.dash')
}

function onDocFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  docFile.value = input.files?.[0] ?? null
  docError.value = ''
}

async function uploadDocument() {
  if (!docFile.value) return
  docBusy.value = true
  docError.value = ''
  try {
    const fd = new FormData()
    fd.append('file', docFile.value)
    if (docTitle.value.trim()) fd.append('title', docTitle.value.trim())
    await $fetch(`/api/pets/${petId}/documents`, { method: 'POST', body: fd })
    docTitle.value = ''
    docFile.value = null
    if (docInputEl.value) docInputEl.value.value = ''
    await loadDocuments()
  } catch (e: any) {
    docError.value = mapError(e)
  } finally {
    docBusy.value = false
  }
}

async function deleteDocument(id: string) {
  docBusy.value = true
  docError.value = ''
  try {
    await $fetch(`/api/pets/documents/${id}`, { method: 'DELETE' })
    await loadDocuments()
  } catch (e: any) {
    docError.value = mapError(e)
  } finally {
    docBusy.value = false
  }
}

async function openMessaging() {
  const ownerId = pet.value?.ownerUserId || clientId
  if (!ownerId) return
  messagingBusy.value = true
  messagingError.value = ''
  try {
    const res: any = await $fetch('/api/messaging/threads', {
      method: 'POST',
      body: { clientUserId: ownerId },
    })
    const thread = res.data ?? res
    await router.push({ path: '/messages', query: { thread: thread.id } })
  } catch (e: any) {
    messagingError.value = mapError(e)
  } finally {
    messagingBusy.value = false
  }
}

async function createCare() {
  careBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/care-reminders`, {
      method: 'POST',
      body: { title: careDraft.title, type: careDraft.type, dueDays: 30 },
    })
    careDraft.title = ''
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function markCareDone(id: string) {
  careBusy.value = true
  try {
    await $fetch(`/api/care-reminders/${id}/done`, { method: 'POST' })
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function postponeCare(id: string, days: number) {
  careBusy.value = true
  try {
    await $fetch(`/api/care-reminders/${id}/postpone`, { method: 'POST', body: { days } })
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function proposeVisit(confirmDirect: boolean) {
  if (!visitDraft.scheduledAt) return
  visitBusy.value = true
  try {
    const body: Record<string, unknown> = {
      notes: visitDraft.notes,
      confirmDirect,
      requestPreconsult: visitDraft.requestPreconsult,
      scheduledAt: new Date(visitDraft.scheduledAt).toISOString(),
    }
    if (visitDraft.visitTypeId) {
      body.visitTypeId = visitDraft.visitTypeId
    } else {
      body.durationMinutes = Number(visitDraft.durationMinutes) || 30
    }
    await $fetch(`/api/pets/${petId}/visits`, {
      method: 'POST',
      body,
    })
    visitDraft.notes = ''
    visitDraft.scheduledAt = ''
    visitDraft.requestPreconsult = false
    visitDraft.visitTypeId = ''
    visitDraft.durationMinutes = 30
    await loadCareAndVisits()
  } finally {
    visitBusy.value = false
  }
}

async function visitAction(id: string, action: string) {
  if (action === 'cancel' && !window.confirm(t('clients.pet.visitCancelConfirm'))) return
  if (action === 'reject_reschedule' && !window.confirm(t('calendar.rejectRescheduleConfirm'))) return
  visitBusy.value = true
  pageError.value = ''
  try {
    const body: Record<string, unknown> = { action }
    if (action === 'confirm' && confirmPreconsultByVisit[id]) {
      body.requestPreconsult = true
    }
    await $fetch(`/api/visits/${id}`, { method: 'PATCH', body })
    if (action === 'confirm') {
      delete confirmPreconsultByVisit[id]
    }
    await loadCareAndVisits()
  } catch (e: any) {
    pageError.value = mapError(e)
  } finally {
    visitBusy.value = false
  }
}

function openVisitReport(v: { id: string, scheduledAt?: string, createdAt?: string }) {
  visitReportId.value = v.id
  visitReportScheduledAt.value = v.scheduledAt || v.createdAt || ''
  visitReportOpen.value = true
}

async function loadDafDispenses() {
  if (!pharmacyEnabled.value || !canReadPharmacy.value) return
  dafDispensesLoading.value = true
  dafDispensesError.value = ''
  try {
    const res: any = await $fetch(`/api/pets/${petId}/daf-dispenses`)
    const data = res?.data ?? res
    dafDispenses.value = Array.isArray(data?.items) ? data.items : []
  }
  catch (e: any) {
    dafDispenses.value = []
    dafDispensesError.value = mapError(e)
  }
  finally {
    dafDispensesLoading.value = false
  }
}

onMounted(async () => {
  pageError.value = ''
  try {
    await fetchUser()
    const petRes: any = await $fetch(`/api/pets/${petId}`)
    pet.value = petRes.data ?? petRes
    petPhotoUrl.value = pet.value?.photoUrl || ''
    horseFoodChainYesNo.value = foodChainToYesNo(pet.value?.foodChainStatus)
    horseDomicile.value = pet.value?.domicileLocation || ''
    lifecycleAdoptedAt.value = toDateInputValue(pet.value?.adoptedAt)
    lifecycleSoldAt.value = toDateInputValue(pet.value?.soldAt)
    lifecycleDeceasedAt.value = toDateInputValue(pet.value?.deceasedAt)

    // Care/visits must not wait on HR sessions — demo pets can have large histories
    // and staging Cloud Run e2e times out waiting for pet-visit-report-open.
    const careVisitsP = loadCareAndVisits().catch((e: any) => {
      pageError.value = mapError(e)
    })
    const docsP = loadDocuments().catch((e: any) => {
      docError.value = mapError(e)
    })
    const sharesP = canReadShares.value
      ? loadPetShares().catch(() => {
          petShares.value = []
        })
      : Promise.resolve()
    const typesP = canManageCalendar.value
      ? $fetch('/api/vet/visit-types?active=1')
          .then((res: any) => {
            const list = res.data ?? res ?? []
            visitTypes.value = (Array.isArray(list) ? list : [])
              .filter((vt: any) => vt?.id)
              .map((vt: any) => ({
                id: vt.id,
                name: vt.name,
                durationMinutes: vt.durationMinutes || 30,
              }))
          })
          .catch(() => {
            visitTypes.value = []
          })
      : Promise.resolve()

    await loadTimeline().catch(() => { timeline.value = [] })
    await loadSessions(true)
    await Promise.all([
      loadWeights().catch(() => { weights.value = [] }),
      loadBloodPressures().catch(() => { bloodPressures.value = [] }),
      loadLabPanels().catch(() => { labPanels.value = [] }),
    ])
    sessionsPollTimer = setInterval(() => {
      loadSessions(true).catch(() => {})
    }, 8000)

    await Promise.all([careVisitsP, docsP, sharesP, typesP, loadDafDispenses()])
  } catch (e: any) {
    pageError.value = mapError(e)
  }
})

onBeforeUnmount(() => {
  if (sessionsPollTimer) clearInterval(sessionsPollTimer)
})
</script>

<style scoped>
.pro-alert-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  margin-bottom: 1rem;
  border-radius: var(--pf-vet-radius);
  border: 1px solid color-mix(in srgb, var(--pf-vet-alert) 35%, transparent);
  background: color-mix(in srgb, var(--pf-vet-alert) 8%, var(--pf-vet-surface));
}

.pro-pet-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
  margin: 0;
}
.pro-pet-horse-reg {
  display: grid;
  gap: 1rem;
}
@media (min-width: 640px) {
  .pro-pet-horse-reg {
    grid-template-columns: 1fr 1fr;
  }
}
.pro-pet-summary dt {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
  margin: 0 0 0.25rem;
}
.pro-pet-summary dd {
  margin: 0;
  font-weight: 600;
}

.pro-pet-filter {
  margin-bottom: 1rem;
}

.pro-pet-reading-date {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pro-pet-reading-comment {
  display: inline-block;
  max-width: 14rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}

.pro-table-row--alert {
  background: color-mix(in srgb, var(--pf-vet-alert) 6%, transparent);
}

.pro-pet-header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.pro-pet-inline-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.pro-pet-inline-form .pro-input {
  flex: 1 1 10rem;
  min-width: 8rem;
}

.pro-pet-charts-preview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
  gap: 0.75rem;
  margin: 0 0 1rem;
  padding: 0;
  list-style: none;
}

.pro-pet-charts-preview li {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-bg, var(--pf-vet-surface));
  font-size: 0.875rem;
  color: var(--pf-vet-text-muted, #64748b);
}

.pro-pet-charts-preview strong {
  color: var(--pf-vet-primary);
  font-variant-numeric: tabular-nums;
}

.pet-daf-dispenses {
  margin: 0;
  padding-left: 1rem;
  display: grid;
  gap: 0.65rem;
}
.pet-daf-dispenses__items {
  margin: 0.35rem 0 0;
  padding-left: 1rem;
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #64748b);
}

.pet-visit-report-modal {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.pet-visit-report-modal__link {
  flex-shrink: 0;
  margin: 0;
}
</style>
