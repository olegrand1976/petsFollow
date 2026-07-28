<template>
  <div
    ref="rootEl"
    class="pf-deck"
    data-testid="commercial-pitch-deck"
    @touchstart.passive="onTouchStart"
    @touchend.passive="onTouchEnd"
  >
    <header class="pf-deck__header">
      <div class="pf-deck__brand">
        <NuxtLink to="/commercial/pitch" class="pf-deck__back" data-testid="pitch-deck-back">
          <ProIcon name="arrow_back" :size="18" />
          <span>{{ t('pitchDeck.ui.back') }}</span>
        </NuxtLink>
        <span class="pf-deck__badge">{{ t('pitchDeck.ui.badge') }}</span>
      </div>

      <div class="pf-deck__counter">
        <span class="pf-deck__muted">{{ t('pitchDeck.ui.slideLabel') }}</span>
        <span class="pf-deck__num">{{ current + 1 }}</span>
        <span class="pf-deck__muted">/ {{ SLIDE_IDS.length }}</span>
        <span class="pf-deck__cat">{{ slideMeta.category }}</span>
      </div>

      <div class="pf-deck__tools">
        <button type="button" class="pf-deck__icon-btn" :title="t('pitchDeck.ui.notesTitle')" data-testid="pitch-deck-notes" @click="notesOpen = !notesOpen">
          <ProIcon name="description" :size="20" />
        </button>
        <button type="button" class="pf-deck__icon-btn" :title="t('pitchDeck.ui.gridTitle')" data-testid="pitch-deck-grid" @click="gridOpen = !gridOpen">
          <ProIcon name="grid_view" :size="20" />
        </button>
        <button type="button" class="pf-deck__icon-btn" :title="t('pitchDeck.ui.fsTitle')" data-testid="pitch-deck-fs" @click="toggleFullscreen">
          <ProIcon name="fullscreen" :size="20" />
        </button>
      </div>
    </header>

    <main class="pf-deck__main">
      <div class="pf-deck__canvas">
        <!-- s01 -->
        <div v-if="slideId === 's01'" class="pf-slide">
          <div class="pf-slide__top">
            <span class="pf-pill pf-pill--teal">{{ s.eyebrow }}</span>
            <span class="pf-mono pf-muted">{{ s.sector }}</span>
          </div>
          <div class="pf-slide__body">
            <div class="pf-pill pf-pill--gold">
              <ProIcon name="verified" :size="16" />
              <span>{{ s.tagline }}</span>
            </div>
            <h1 class="pf-slide__hero">
              {{ s.headline }} <span class="pf-teal">{{ t('pitchDeck.ui.brandSuffix') }}</span>
            </h1>
            <p class="pf-slide__lead">{{ s.lead }}</p>
            <div class="pf-grid-3">
              <div class="pf-card">
                <div class="pf-card__title"><ProIcon name="computer" :size="18" class="pf-teal" />{{ s.vetProTitle }}</div>
                <p>{{ s.vetProDesc }}</p>
              </div>
              <div class="pf-card">
                <div class="pf-card__title"><ProIcon name="directions_car" :size="18" class="pf-teal" />{{ s.vetLightTitle }}</div>
                <p>{{ s.vetLightDesc }}</p>
              </div>
              <div class="pf-card">
                <div class="pf-card__title"><ProIcon name="devices" :size="18" class="pf-teal" />{{ s.clientTitle }}</div>
                <p>{{ s.clientDesc }}</p>
              </div>
            </div>
          </div>
          <div class="pf-slide__foot">
            <span>{{ s.footerLeft }}</span>
            <span class="pf-teal pf-bold">{{ s.footerRight }}</span>
          </div>
        </div>

        <!-- s02 -->
        <div v-else-if="slideId === 's02'" class="pf-slide">
          <div>
            <span class="pf-kicker pf-kicker--alert">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-3 pf-slide__grow">
            <div class="pf-card pf-card--elev">
              <div class="pf-icon-circle pf-icon-circle--alert"><ProIcon name="comments_disabled" :size="20" /></div>
              <h3>{{ s.card1Title }}</h3>
              <p>{{ s.card1Desc }}</p>
            </div>
            <div class="pf-card pf-card--elev">
              <div class="pf-icon-circle pf-icon-circle--coral"><ProIcon name="sms_failed" :size="20" /></div>
              <h3>{{ s.card2Title }}</h3>
              <p>{{ s.card2Desc }}</p>
            </div>
            <div class="pf-card pf-card--elev">
              <div class="pf-icon-circle pf-icon-circle--gold"><ProIcon name="app_badging" :size="20" /></div>
              <h3>{{ s.card3Title }}</h3>
              <p>{{ s.card3Desc }}</p>
            </div>
          </div>
          <div class="pf-banner pf-banner--teal">
            <ProIcon name="check_circle" :size="18" />
            <span>{{ s.solution }}</span>
          </div>
        </div>

        <!-- s03 -->
        <div v-else-if="slideId === 's03'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-3 pf-slide__grow">
            <div class="pf-card pf-card--navy">
              <div class="pf-card__meta">
                <span class="pf-tag">{{ s.vetProLabel }}</span>
                <span class="pf-muted-on-dark">{{ s.vetProKind }}</span>
              </div>
              <h3>{{ s.vetProTitle }}</h3>
              <ul>
                <li v-for="item in list('vetProItems')" :key="item"><ProIcon name="check" :size="14" />{{ item }}</li>
              </ul>
              <div class="pf-card__price pf-gold">{{ s.vetProPrice }}</div>
            </div>
            <div class="pf-card pf-card--border-teal">
              <div class="pf-card__meta">
                <span class="pf-tag pf-tag--outline">{{ s.vetLightLabel }}</span>
                <span class="pf-muted">{{ s.vetLightKind }}</span>
              </div>
              <h3>{{ s.vetLightTitle }}</h3>
              <ul class="pf-list-dark">
                <li v-for="item in list('vetLightItems')" :key="item"><ProIcon name="check" :size="14" />{{ item }}</li>
              </ul>
              <div class="pf-card__price pf-sage">{{ s.vetLightPrice }}</div>
            </div>
            <div class="pf-card pf-card--elev">
              <div class="pf-card__meta">
                <span class="pf-tag pf-tag--sage">{{ s.clientLabel }}</span>
                <span class="pf-muted">{{ s.clientKind }}</span>
              </div>
              <h3>{{ s.clientTitle }}</h3>
              <ul class="pf-list-dark">
                <li v-for="item in list('clientItems')" :key="item"><ProIcon name="check" :size="14" />{{ item }}</li>
              </ul>
              <div class="pf-card__price pf-teal">{{ s.clientPrice }}</div>
            </div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s04 -->
        <div v-else-if="slideId === 's04'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-eco pf-slide__grow">
            <div class="pf-grid-3 pf-eco__row">
              <div class="pf-card pf-card--navy pf-center">
                <div class="pf-icon-circle pf-icon-circle--teal-solid"><ProIcon name="computer" :size="22" /></div>
                <h4>{{ s.vetProTitle }}</h4>
                <p class="pf-muted-on-dark">{{ s.vetProDesc }}</p>
                <span class="pf-chip">{{ s.vetProBadge }}</span>
              </div>
              <div class="pf-eco__flow">
                <span class="pf-pill pf-pill--teal">{{ s.flowLabel }} <ProIcon name="arrow_forward" :size="14" /></span>
                <div class="pf-eco__line" />
                <span class="pf-muted pf-small">{{ s.flowHint }}</span>
              </div>
              <div class="pf-card pf-card--elev pf-center pf-card--border-sage">
                <div class="pf-icon-circle pf-icon-circle--sage-solid"><ProIcon name="devices" :size="22" /></div>
                <h4>{{ s.clientTitle }}</h4>
                <p class="pf-muted">{{ s.clientDesc }}</p>
                <span class="pf-chip pf-chip--sage">{{ s.clientBadge }}</span>
              </div>
            </div>
            <div class="pf-eco__sub">
              <div class="pf-eco__sub-left">
                <div class="pf-icon-circle pf-icon-circle--teal"><ProIcon name="directions_car" :size="18" /></div>
                <div>
                  <strong>{{ s.vetLightLine }}</strong>
                  <span class="pf-muted"> {{ s.vetLightDesc }}</span>
                </div>
              </div>
              <span class="pf-chip pf-chip--teal">{{ s.syncBadge }}</span>
            </div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s05 -->
        <div v-else-if="slideId === 's05'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-2 pf-slide__grow">
            <div class="pf-stack">
              <div class="pf-card pf-card--navy">
                <div class="pf-card__title pf-gold"><ProIcon name="support_agent" :size="16" />{{ s.heroTitle }}</div>
                <p>{{ s.heroDesc }}</p>
              </div>
              <div class="pf-point"><ProIcon name="check_circle" :size="18" class="pf-teal" /><div><strong>{{ s.point1Title }}</strong><span class="pf-muted"> {{ s.point1Desc }}</span></div></div>
              <div class="pf-point"><ProIcon name="check_circle" :size="18" class="pf-teal" /><div><strong>{{ s.point2Title }}</strong><span class="pf-muted"> {{ s.point2Desc }}</span></div></div>
              <div class="pf-point"><ProIcon name="check_circle" :size="18" class="pf-teal" /><div><strong>{{ s.point3Title }}</strong><span class="pf-muted"> {{ s.point3Desc }}</span></div></div>
            </div>
            <div class="pf-mock">
              <div class="pf-mock__head">
                <span class="pf-bold">{{ s.mockTitle }}</span>
                <span class="pf-chip pf-chip--teal">{{ s.mockAccess }}</span>
              </div>
              <div class="pf-grid-2">
                <div class="pf-card"><span class="pf-muted pf-tiny">{{ s.mockPatients }}</span><strong class="pf-mono">128</strong></div>
                <div class="pf-card"><span class="pf-muted pf-tiny">{{ s.mockCommissions }}</span><strong class="pf-mono pf-teal">245 €</strong></div>
              </div>
              <div class="pf-card pf-row">
                <span class="pf-row"><ProIcon name="forum" :size="16" class="pf-teal" /><strong>{{ s.mockMessage }}</strong></span>
                <span class="pf-chip pf-chip--sage">{{ s.mockStatus }}</span>
              </div>
            </div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s06 -->
        <div v-else-if="slideId === 's06'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-2 pf-slide__grow">
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-feat__icon"><ProIcon name="forum" :size="20" /></div><div><h4>{{ s.f1Title }}</h4><p>{{ s.f1Desc }}</p></div></div>
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-feat__icon"><ProIcon name="timeline" :size="20" /></div><div><h4>{{ s.f2Title }}</h4><p>{{ s.f2Desc }}</p></div></div>
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-feat__icon"><ProIcon name="description" :size="20" /></div><div><h4>{{ s.f3Title }}</h4><p>{{ s.f3Desc }}</p></div></div>
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-feat__icon"><ProIcon name="group_add" :size="20" /></div><div><h4>{{ s.f4Title }}</h4><p>{{ s.f4Desc }}</p></div></div>
          </div>
          <div class="pf-banner pf-banner--gold">
            <span class="pf-row"><ProIcon name="workspace_premium" :size="16" /><span>{{ s.quality }}</span></span>
            <span class="pf-mono pf-bold">{{ s.langs }}</span>
          </div>
        </div>

        <!-- s07 -->
        <div v-else-if="slideId === 's07'" class="pf-slide">
          <div>
            <span class="pf-kicker pf-kicker--sage">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-2 pf-slide__grow">
            <div class="pf-stack">
              <div class="pf-card pf-card--teal">
                <div class="pf-card__meta">
                  <span class="pf-gold pf-tiny pf-mono pf-bold">{{ s.freeBadge }}</span>
                  <span class="pf-tiny pf-mono">{{ s.mobileBadge }}</span>
                </div>
                <h3>{{ s.heroTitle }}</h3>
                <p>{{ s.heroDesc }}</p>
              </div>
              <div class="pf-point"><ProIcon name="check_circle" :size="16" class="pf-teal" /><span>{{ s.point1 }}</span></div>
              <div class="pf-point"><ProIcon name="check_circle" :size="16" class="pf-teal" /><span>{{ s.point2 }}</span></div>
              <div class="pf-point"><ProIcon name="check_circle" :size="16" class="pf-teal" /><span>{{ s.point3 }}</span></div>
            </div>
            <div class="pf-phone">
              <div class="pf-phone__bar"><span>09:41</span><span class="pf-sage">VetLight Mobile</span><ProIcon name="battery_full" :size="12" /></div>
              <div class="pf-phone__panel">
                <span class="pf-tiny pf-teal pf-mono pf-bold">{{ s.phoneAgenda }}</span>
                <div class="pf-phone__event">
                  <div><strong>{{ s.phoneEvent }}</strong><span class="pf-muted">{{ s.phoneEventDesc }}</span></div>
                  <span class="pf-chip pf-chip--sage">{{ s.phoneDone }}</span>
                </div>
              </div>
              <div class="pf-phone__cta"><ProIcon name="mic" :size="16" />{{ s.phoneCta }}</div>
            </div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s08 -->
        <div v-else-if="slideId === 's08'" class="pf-slide">
          <div>
            <span class="pf-kicker pf-kicker--sage">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-3 pf-slide__grow">
            <div class="pf-card pf-card--elev">
              <div class="pf-icon-circle pf-icon-circle--sage"><ProIcon name="forum" :size="18" /></div>
              <h3>{{ s.c1Title }}</h3><p>{{ s.c1Desc }}</p>
            </div>
            <div class="pf-card pf-card--elev">
              <div class="pf-icon-circle pf-icon-circle--teal"><ProIcon name="all_inclusive" :size="18" /></div>
              <h3>{{ s.c2Title }}</h3><p>{{ s.c2Desc }}</p>
            </div>
            <div class="pf-card pf-card--elev">
              <div class="pf-icon-circle pf-icon-circle--gold"><ProIcon name="favorite" :size="18" /></div>
              <h3>{{ s.c3Title }}</h3><p>{{ s.c3Desc }}</p>
            </div>
          </div>
          <div class="pf-banner pf-banner--sage">
            <span>{{ s.footerLeft }}</span>
            <span class="pf-mono pf-bold pf-sage">{{ s.footerRight }}</span>
          </div>
        </div>

        <!-- s09 -->
        <div v-else-if="slideId === 's09'" class="pf-slide">
          <div>
            <span class="pf-kicker pf-kicker--gold">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-3 pf-slide__grow">
            <div class="pf-card pf-card--elev pf-plan">
              <span class="pf-muted pf-mono pf-tiny pf-bold">{{ s.monthlyName }}</span>
              <div class="pf-plan__price"><strong>{{ s.monthlyPrice }}</strong><span class="pf-muted">{{ s.monthlyPeriod }}</span></div>
              <p class="pf-muted">{{ s.monthlyDesc }}</p>
              <div class="pf-plan__note">{{ s.monthlyNote }}</div>
            </div>
            <div class="pf-card pf-card--elev pf-plan">
              <span class="pf-teal pf-mono pf-tiny pf-bold">{{ s.annualName }}</span>
              <div class="pf-plan__price"><strong>{{ s.annualPrice }}</strong><span class="pf-muted">{{ s.annualPeriod }}</span></div>
              <p class="pf-muted">{{ s.annualDesc }}</p>
              <div class="pf-plan__note pf-teal">{{ s.annualNote }}</div>
            </div>
            <div class="pf-card pf-card--navy pf-plan pf-plan--rec">
              <span class="pf-plan__rec">{{ s.recommended }}</span>
              <span class="pf-gold pf-mono pf-tiny pf-bold">{{ s.triennialName }}</span>
              <div class="pf-plan__price"><strong>{{ s.triennialPrice }}</strong><span>{{ s.triennialPeriod }}</span></div>
              <p>{{ s.triennialDesc }}</p>
              <div class="pf-plan__note pf-gold">{{ s.triennialNote }}</div>
            </div>
          </div>
          <div class="pf-banner pf-banner--teal">
            <ProIcon name="check_circle" :size="16" />
            <span>{{ s.included }}</span>
          </div>
        </div>

        <!-- s10 -->
        <div v-else-if="slideId === 's10'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-table-wrap pf-slide__grow">
            <table class="pf-table">
              <thead>
                <tr>
                  <th>{{ s.colFeature }}</th>
                  <th>{{ s.colVetPro }}</th>
                  <th>{{ s.colVetLight }}</th>
                  <th>{{ s.colClient }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, i) in matrixRows" :key="i">
                  <td>{{ row.feature }}</td>
                  <td class="pf-center" :class="cellClass(row.vetPro)">{{ row.vetPro }}</td>
                  <td class="pf-center" :class="cellClass(row.vetLight)">{{ row.vetLight }}</td>
                  <td class="pf-center" :class="cellClass(row.client)">{{ row.client }}</td>
                </tr>
                <tr class="pf-table__finance">
                  <td>{{ s.financeLabel }}</td>
                  <td class="pf-center pf-bold">{{ s.financeVetPro }}</td>
                  <td class="pf-center pf-bold pf-teal">{{ s.financeVetLight }}</td>
                  <td class="pf-center pf-bold pf-sage">{{ s.financeClient }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s11 -->
        <div v-else-if="slideId === 's11'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-2 pf-slide__grow">
            <div class="pf-stack">
              <div class="pf-card pf-card--navy">
                <span class="pf-gold pf-mono pf-tiny pf-bold">{{ s.priceLabel }}</span>
                <div class="pf-plan__price"><strong>{{ s.price }}</strong></div>
                <p>{{ s.setup }}</p>
              </div>
              <div class="pf-card pf-card--elev">
                <strong class="pf-teal">{{ s.mechanismTitle }}</strong>
                <p class="pf-muted">{{ s.mechanismDesc }}</p>
              </div>
            </div>
            <div class="pf-mock">
              <strong>{{ s.summaryTitle }}</strong>
              <div v-for="item in list('summaryItems')" :key="item" class="pf-point">
                <ProIcon name="check_circle" :size="18" class="pf-teal" /><span>{{ item }}</span>
              </div>
            </div>
          </div>
          <div class="pf-banner pf-banner--navy">{{ s.footer }}</div>
        </div>

        <!-- s12 -->
        <div v-else-if="slideId === 's12'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-2 pf-slide__grow">
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-icon-circle pf-icon-circle--teal"><ProIcon name="devices" :size="22" /></div><div><h4>{{ s.d1Title }}</h4><p>{{ s.d1Desc }}</p></div></div>
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-icon-circle pf-icon-circle--teal"><ProIcon name="view_module" :size="22" /></div><div><h4>{{ s.d2Title }}</h4><p>{{ s.d2Desc }}</p></div></div>
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-icon-circle pf-icon-circle--teal"><ProIcon name="directions_car" :size="22" /></div><div><h4>{{ s.d3Title }}</h4><p>{{ s.d3Desc }}</p></div></div>
            <div class="pf-card pf-card--elev pf-feat"><div class="pf-icon-circle pf-icon-circle--teal"><ProIcon name="all_inclusive" :size="22" /></div><div><h4>{{ s.d4Title }}</h4><p>{{ s.d4Desc }}</p></div></div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s13 -->
        <div v-else-if="slideId === 's13'" class="pf-slide">
          <div>
            <span class="pf-kicker pf-kicker--gold">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-grid-2 pf-slide__grow">
            <div v-for="(obj, i) in objections" :key="i" class="pf-card pf-card--elev">
              <span class="pf-tiny pf-bold pf-alert pf-mono">{{ obj.label }}</span>
              <h4 class="pf-mono">{{ obj.q }}</h4>
              <p class="pf-muted"><strong class="pf-teal">{{ s.answerLabel }}</strong> {{ obj.a }}</p>
            </div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s14 -->
        <div v-else-if="slideId === 's14'" class="pf-slide">
          <div>
            <span class="pf-kicker">{{ s.kicker }}</span>
            <h2 class="pf-slide__h2">{{ s.headline }}</h2>
          </div>
          <div class="pf-demo pf-slide__grow">
            <div class="pf-demo__tabs">
              <button
                v-for="(step, i) in demoSteps"
                :key="i"
                type="button"
                class="pf-demo__tab"
                :class="{ 'pf-demo__tab--active': demoStep === i }"
                :data-testid="`pitch-deck-demo-${i}`"
                @click="demoStep = i"
              >
                {{ step.tab }}
              </button>
            </div>
            <div class="pf-card pf-card--elev pf-demo__content" data-testid="pitch-deck-demo-content">
              <ProIcon :name="demoSteps[demoStep]?.icon || 'computer'" :size="28" class="pf-teal" />
              <div>
                <h4>{{ demoSteps[demoStep]?.title }}</h4>
                <p class="pf-muted">{{ demoSteps[demoStep]?.desc }}</p>
              </div>
            </div>
          </div>
          <p class="pf-muted pf-small">{{ s.footer }}</p>
        </div>

        <!-- s15 -->
        <div v-else-if="slideId === 's15'" class="pf-slide">
          <div class="pf-slide__top">
            <span class="pf-kicker">{{ s.kicker }}</span>
            <span class="pf-pill pf-pill--gold">{{ s.badge }}</span>
          </div>
          <div class="pf-slide__body pf-center">
            <h1 class="pf-slide__hero pf-slide__hero--sm">{{ s.headline }}</h1>
            <div class="pf-grid-3">
              <div class="pf-card pf-card--elev pf-left">
                <div class="pf-card__title pf-teal"><ProIcon name="contact_phone" :size="16" />{{ s.c1Title }}</div>
                <p>{{ s.c1Desc }}</p>
              </div>
              <div class="pf-card pf-card--elev pf-left">
                <div class="pf-card__title pf-teal"><ProIcon name="bolt" :size="16" />{{ s.c2Title }}</div>
                <p>{{ s.c2Desc }}</p>
              </div>
              <div class="pf-card pf-card--elev pf-left">
                <div class="pf-card__title pf-teal"><ProIcon name="card_giftcard" :size="16" />{{ s.c3Title }}</div>
                <p>{{ s.c3Desc }}</p>
              </div>
            </div>
            <button type="button" class="pf-cta" data-testid="pitch-deck-cta">{{ s.cta }}</button>
          </div>
          <div class="pf-slide__foot">
            <span>{{ s.footerLeft }}</span>
            <span class="pf-bold">{{ s.footerRight }}</span>
          </div>
        </div>
      </div>

      <aside v-if="notesOpen" class="pf-notes" data-testid="pitch-deck-notes-drawer">
        <div class="pf-notes__head">
          <span class="pf-row"><ProIcon name="record_voice_over" :size="14" /><span>{{ t('pitchDeck.ui.notesTitle') }}</span></span>
          <button type="button" class="pf-deck__icon-btn" @click="notesOpen = false"><ProIcon name="close" :size="16" /></button>
        </div>
        <p>{{ slideMeta.notes }}</p>
      </aside>

      <div v-if="gridOpen" class="pf-grid-modal" data-testid="pitch-deck-grid-modal" @click.self="gridOpen = false">
        <div class="pf-grid-modal__panel">
          <div class="pf-grid-modal__head">
            <div>
              <h3>{{ t('pitchDeck.ui.gridTitle') }}</h3>
              <p>{{ t('pitchDeck.ui.gridHint') }}</p>
            </div>
            <button type="button" class="pf-deck__icon-btn" @click="gridOpen = false"><ProIcon name="close" :size="20" /></button>
          </div>
          <div class="pf-grid-modal__list">
            <button
              v-for="(id, idx) in SLIDE_IDS"
              :key="id"
              type="button"
              class="pf-grid-modal__item"
              :class="{ 'pf-grid-modal__item--active': idx === current }"
              @click="jumpTo(idx)"
            >
              <div class="pf-grid-modal__meta">
                <span>{{ t('pitchDeck.ui.slideLabel') }} {{ idx + 1 }}</span>
                <span class="pf-teal">{{ metaFor(id).category }}</span>
              </div>
              <strong>{{ metaFor(id).title }}</strong>
            </button>
          </div>
        </div>
      </div>
    </main>

    <footer class="pf-deck__footer">
      <div class="pf-deck__shortcuts">
        <span class="pf-muted">{{ t('pitchDeck.ui.shortcuts') }}</span>
        <kbd>←</kbd><kbd>→</kbd>
      </div>
      <div class="pf-deck__nav">
        <button type="button" class="pf-btn pf-btn--ghost" :disabled="current === 0" data-testid="pitch-deck-prev" @click="prev">
          <ProIcon name="arrow_back" :size="16" />{{ t('pitchDeck.ui.prev') }}
        </button>
        <button type="button" class="pf-btn pf-btn--primary" :disabled="current === SLIDE_IDS.length - 1" data-testid="pitch-deck-next" @click="next">
          {{ t('pitchDeck.ui.next') }}<ProIcon name="arrow_forward" :size="16" />
        </button>
      </div>
      <div class="pf-bold pf-mono pf-small">{{ t('pitchDeck.ui.footerBrand') }}</div>
    </footer>
  </div>
</template>

<script setup lang="ts">
const { t, tm, rt } = useI18n()

const SLIDE_IDS = [
  's01', 's02', 's03', 's04', 's05', 's06', 's07', 's08', 's09', 's10', 's11', 's12', 's13', 's14', 's15',
] as const

type SlideId = typeof SLIDE_IDS[number]

const current = ref(0)
const notesOpen = ref(false)
const gridOpen = ref(false)
const demoStep = ref(0)
const rootEl = ref<HTMLElement | null>(null)
let touchStartX = 0

const slideId = computed(() => SLIDE_IDS[current.value])

const s = computed(() => {
  const raw = tm(`pitchDeck.slides.${slideId.value}`) as Record<string, unknown>
  return new Proxy({} as Record<string, string>, {
    get(_target, prop: string) {
      const val = raw?.[prop]
      if (typeof val === 'string') return val
      if (val && typeof val === 'object' && 'loc' in (val as object)) return rt(val as any)
      return ''
    },
  })
})

function metaFor(id: SlideId) {
  return {
    title: t(`pitchDeck.slides.${id}.title`),
    category: t(`pitchDeck.slides.${id}.category`),
    notes: t(`pitchDeck.slides.${id}.notes`),
  }
}

const slideMeta = computed(() => metaFor(slideId.value))

function list(key: string): string[] {
  const raw = tm(`pitchDeck.slides.${slideId.value}.${key}`) as unknown
  if (!Array.isArray(raw)) return []
  return raw.map((x) => (typeof x === 'string' ? x : rt(x as any)))
}

type MatrixRow = { feature: string; vetPro: string; vetLight: string; client: string }
const matrixRows = computed(() => {
  const raw = tm('pitchDeck.slides.s10.rows') as unknown
  if (!Array.isArray(raw)) return [] as MatrixRow[]
  return raw.map((row: any) => ({
    feature: typeof row.feature === 'string' ? row.feature : rt(row.feature),
    vetPro: typeof row.vetPro === 'string' ? row.vetPro : rt(row.vetPro),
    vetLight: typeof row.vetLight === 'string' ? row.vetLight : rt(row.vetLight),
    client: typeof row.client === 'string' ? row.client : rt(row.client),
  }))
})

type Objection = { label: string; q: string; a: string }
const objections = computed(() => {
  const raw = tm('pitchDeck.slides.s13.objections') as unknown
  if (!Array.isArray(raw)) return [] as Objection[]
  return raw.map((o: any) => ({
    label: typeof o.label === 'string' ? o.label : rt(o.label),
    q: typeof o.q === 'string' ? o.q : rt(o.q),
    a: typeof o.a === 'string' ? o.a : rt(o.a),
  }))
})

type DemoStep = { tab: string; icon: string; title: string; desc: string }
const demoSteps = computed(() => {
  const raw = tm('pitchDeck.slides.s14.steps') as unknown
  if (!Array.isArray(raw)) return [] as DemoStep[]
  return raw.map((st: any) => ({
    tab: typeof st.tab === 'string' ? st.tab : rt(st.tab),
    icon: typeof st.icon === 'string' ? st.icon : 'computer',
    title: typeof st.title === 'string' ? st.title : rt(st.title),
    desc: typeof st.desc === 'string' ? st.desc : rt(st.desc),
  }))
})

function cellClass(val: string) {
  if (val.startsWith('✓')) return 'pf-teal pf-bold'
  if (val === '—') return 'pf-muted'
  return ''
}

function next() {
  if (current.value < SLIDE_IDS.length - 1) current.value += 1
}
function prev() {
  if (current.value > 0) current.value -= 1
}
function jumpTo(idx: number) {
  current.value = idx
  gridOpen.value = false
}

function toggleFullscreen() {
  const el = rootEl.value
  if (!el) return
  if (!document.fullscreenElement) {
    void el.requestFullscreen?.().catch(() => {})
  } else {
    void document.exitFullscreen?.()
  }
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'ArrowRight' || e.key === ' ') {
    e.preventDefault()
    next()
  } else if (e.key === 'ArrowLeft') {
    prev()
  } else if (e.key === 'n' || e.key === 'N') {
    notesOpen.value = !notesOpen.value
  } else if (e.key === 'g' || e.key === 'G') {
    gridOpen.value = !gridOpen.value
  } else if (e.key === 'f' || e.key === 'F') {
    toggleFullscreen()
  } else if (e.key === 'Escape') {
    notesOpen.value = false
    gridOpen.value = false
  }
}

function onTouchStart(e: TouchEvent) {
  touchStartX = e.touches[0]?.clientX ?? 0
}
function onTouchEnd(e: TouchEvent) {
  const endX = e.changedTouches[0]?.clientX ?? 0
  const delta = touchStartX - endX
  if (delta > 50) next()
  else if (delta < -50) prev()
}

watch(slideId, (id) => {
  if (id === 's14') demoStep.value = 0
})

onMounted(() => {
  window.addEventListener('keydown', onKey)
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKey)
})
</script>

<style scoped>
.pf-deck {
  --deck-navy: var(--pf-vet-primary, #1B3A4B);
  --deck-teal: var(--pf-vet-accent, #2A9D8F);
  --deck-gold: var(--pf-brand-gold, #E9C46A);
  --deck-sage: var(--pf-brand-sage, #52B788);
  --deck-coral: var(--pf-brand-coral, #F4A261);
  --deck-alert: var(--pf-vet-alert, #E76F51);
  --deck-bg: var(--pf-vet-bg, #F7F9FB);
  --deck-surface: var(--pf-vet-surface, #FFFFFF);
  --deck-border: var(--pf-vet-border, #E2E6ED);
  --deck-muted: var(--pf-vet-text-muted, #6B7280);
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  background: var(--deck-bg);
  color: var(--deck-navy);
  font-family: var(--pf-font, 'DM Sans', system-ui, sans-serif);
  overflow: hidden;
}
.pf-deck__header,
.pf-deck__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  height: 3.5rem;
  padding: 0 1rem;
  background: var(--deck-surface);
  border-color: var(--deck-border);
  flex-shrink: 0;
  z-index: 20;
}
.pf-deck__header { border-bottom: 1px solid var(--deck-border); }
.pf-deck__footer { border-top: 1px solid var(--deck-border); }
.pf-deck__brand { display: flex; align-items: center; gap: 0.75rem; min-width: 0; }
.pf-deck__back {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--deck-navy);
  text-decoration: none;
  font-size: 0.8rem;
  font-weight: 600;
}
.pf-deck__back:hover { color: var(--deck-teal); }
.pf-deck__badge {
  display: none;
  font-size: 0.7rem;
  font-family: var(--pf-font-mono, monospace);
  font-weight: 600;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--deck-teal) 12%, white);
  color: var(--deck-teal);
  border: 1px solid color-mix(in srgb, var(--deck-teal) 25%, white);
}
@media (min-width: 720px) { .pf-deck__badge { display: inline-block; } }
.pf-deck__counter {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-family: var(--pf-font-mono, monospace);
  font-weight: 600;
}
.pf-deck__num {
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: var(--deck-navy);
  color: #fff;
}
.pf-deck__cat {
  display: none;
  margin-left: 0.5rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: var(--deck-bg);
  border: 1px solid var(--deck-border);
  color: var(--deck-muted);
  font-size: 0.68rem;
}
@media (min-width: 900px) { .pf-deck__cat { display: inline-block; } }
.pf-deck__tools { display: flex; gap: 0.25rem; }
.pf-deck__icon-btn {
  border: 1px solid transparent;
  background: transparent;
  border-radius: 8px;
  padding: 0.4rem;
  cursor: pointer;
  color: var(--deck-navy);
}
.pf-deck__icon-btn:hover { background: var(--deck-bg); border-color: var(--deck-border); }
.pf-deck__main {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.75rem 1rem;
  overflow: hidden;
}
.pf-deck__canvas {
  width: 100%;
  max-width: 72rem;
  aspect-ratio: 16 / 9;
  max-height: calc(100dvh - 7.5rem);
  background: var(--deck-surface);
  border: 1px solid var(--deck-border);
  border-radius: var(--pf-vet-radius-xl, 16px);
  box-shadow: var(--pf-vet-shadow-md);
  padding: 1.25rem 1.5rem;
  overflow: auto;
}
.pf-slide {
  height: 100%;
  min-height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 0.75rem;
}
.pf-slide__grow { flex: 1; margin: auto 0; }
.pf-slide__top { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; border-bottom: 1px solid var(--deck-border); padding-bottom: 0.65rem; }
.pf-slide__body { display: grid; gap: 0.85rem; margin: auto 0; }
.pf-slide__foot { display: flex; justify-content: space-between; gap: 0.5rem; font-size: 0.7rem; color: var(--deck-muted); font-family: var(--pf-font-mono, monospace); border-top: 1px solid color-mix(in srgb, var(--deck-border) 60%, transparent); padding-top: 0.5rem; }
.pf-slide__hero { margin: 0; font-size: clamp(1.75rem, 4vw, 3rem); font-weight: 800; letter-spacing: -0.02em; line-height: 1.1; }
.pf-slide__hero--sm { font-size: clamp(1.4rem, 3vw, 2.25rem); }
.pf-slide__lead { margin: 0; font-size: clamp(0.95rem, 1.6vw, 1.2rem); font-weight: 600; color: var(--deck-teal); max-width: 48rem; }
.pf-slide__h2 { margin: 0.25rem 0 0; font-size: clamp(1.15rem, 2.2vw, 1.6rem); font-weight: 700; }
.pf-kicker { font-size: 0.7rem; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; font-family: var(--pf-font-mono, monospace); color: var(--deck-teal); }
.pf-kicker--alert { color: var(--deck-alert); }
.pf-kicker--sage { color: var(--deck-sage); }
.pf-kicker--gold { color: var(--deck-gold); }
.pf-grid-2 { display: grid; grid-template-columns: 1fr; gap: 0.85rem; }
.pf-grid-3 { display: grid; grid-template-columns: 1fr; gap: 0.75rem; }
@media (min-width: 800px) {
  .pf-grid-2 { grid-template-columns: 1fr 1fr; }
  .pf-grid-3 { grid-template-columns: repeat(3, 1fr); }
}
.pf-card {
  background: var(--deck-bg);
  border: 1px solid var(--deck-border);
  border-radius: 12px;
  padding: 0.85rem;
}
.pf-card--elev { background: var(--deck-surface); box-shadow: var(--pf-vet-shadow-sm); }
.pf-card--navy { background: var(--deck-navy); color: #fff; border-color: transparent; }
.pf-card--teal { background: var(--deck-teal); color: #fff; border-color: transparent; }
.pf-card--border-teal { border: 2px solid var(--deck-teal); background: var(--deck-surface); }
.pf-card--border-sage { border: 2px solid var(--deck-sage); }
.pf-card h3, .pf-card h4 { margin: 0 0 0.35rem; font-size: 0.9rem; }
.pf-card p { margin: 0; font-size: 0.72rem; color: var(--deck-muted); line-height: 1.45; }
.pf-card--navy p, .pf-card--teal p { color: rgba(255,255,255,0.85); }
.pf-card__title { display: flex; align-items: center; gap: 0.4rem; font-weight: 700; font-size: 0.78rem; margin-bottom: 0.35rem; color: var(--deck-navy); }
.pf-card--navy .pf-card__title, .pf-card--teal .pf-card__title { color: inherit; }
.pf-card__meta { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; gap: 0.5rem; }
.pf-card__price { margin-top: 0.75rem; padding-top: 0.5rem; border-top: 1px solid color-mix(in srgb, var(--deck-border) 70%, transparent); font-size: 0.75rem; font-weight: 700; }
.pf-card--navy .pf-card__price { border-top-color: rgba(255,255,255,0.15); }
.pf-card ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.35rem; font-size: 0.72rem; }
.pf-card ul li { display: flex; align-items: center; gap: 0.35rem; }
.pf-list-dark { color: var(--deck-navy); }
.pf-tag { padding: 0.15rem 0.4rem; border-radius: 4px; background: var(--deck-teal); color: #fff; font-size: 0.65rem; font-weight: 700; font-family: var(--pf-font-mono, monospace); }
.pf-tag--outline { background: color-mix(in srgb, var(--deck-teal) 12%, white); color: var(--deck-teal); border: 1px solid color-mix(in srgb, var(--deck-teal) 30%, white); }
.pf-tag--sage { background: color-mix(in srgb, var(--deck-sage) 12%, white); color: var(--deck-sage); border: 1px solid color-mix(in srgb, var(--deck-sage) 30%, white); }
.pf-pill { display: inline-flex; align-items: center; gap: 0.35rem; padding: 0.25rem 0.65rem; border-radius: 999px; font-size: 0.72rem; font-weight: 700; }
.pf-pill--teal { background: color-mix(in srgb, var(--deck-teal) 12%, white); color: var(--deck-teal); border: 1px solid color-mix(in srgb, var(--deck-teal) 30%, white); }
.pf-pill--gold { background: color-mix(in srgb, var(--deck-gold) 18%, white); color: var(--deck-navy); border: 1px solid color-mix(in srgb, var(--deck-gold) 35%, white); width: fit-content; }
.pf-icon-circle {
  width: 2.1rem; height: 2.1rem; border-radius: 999px; display: inline-flex; align-items: center; justify-content: center; margin-bottom: 0.45rem;
}
.pf-icon-circle--alert { background: color-mix(in srgb, var(--deck-alert) 12%, white); color: var(--deck-alert); }
.pf-icon-circle--coral { background: color-mix(in srgb, var(--deck-coral) 15%, white); color: var(--deck-coral); }
.pf-icon-circle--gold { background: color-mix(in srgb, var(--deck-gold) 18%, white); color: var(--deck-navy); }
.pf-icon-circle--teal { background: color-mix(in srgb, var(--deck-teal) 12%, white); color: var(--deck-teal); }
.pf-icon-circle--sage { background: color-mix(in srgb, var(--deck-sage) 15%, white); color: var(--deck-sage); }
.pf-icon-circle--teal-solid { background: var(--deck-teal); color: #fff; margin-inline: auto; }
.pf-icon-circle--sage-solid { background: var(--deck-sage); color: #fff; margin-inline: auto; }
.pf-banner {
  display: flex; align-items: center; justify-content: space-between; gap: 0.75rem;
  padding: 0.65rem 0.85rem; border-radius: 8px; font-size: 0.75rem;
}
.pf-banner--teal { background: color-mix(in srgb, var(--deck-teal) 10%, white); border: 1px solid color-mix(in srgb, var(--deck-teal) 30%, white); }
.pf-banner--gold { background: color-mix(in srgb, var(--deck-gold) 15%, white); border: 1px solid color-mix(in srgb, var(--deck-gold) 35%, white); }
.pf-banner--sage { background: color-mix(in srgb, var(--deck-sage) 10%, white); border: 1px solid color-mix(in srgb, var(--deck-sage) 30%, white); }
.pf-banner--navy { background: var(--deck-navy); color: #fff; font-family: var(--pf-font-mono, monospace); }
.pf-stack { display: grid; gap: 0.65rem; }
.pf-point { display: flex; align-items: flex-start; gap: 0.45rem; font-size: 0.75rem; }
.pf-feat { display: flex; gap: 0.75rem; align-items: flex-start; }
.pf-feat__icon { padding: 0.45rem; border-radius: 8px; background: color-mix(in srgb, var(--deck-teal) 12%, white); color: var(--deck-teal); }
.pf-mock { display: grid; gap: 0.65rem; padding: 1rem; border-radius: 16px; background: var(--deck-bg); border: 1px solid var(--deck-border); box-shadow: var(--pf-vet-shadow-md); }
.pf-mock__head { display: flex; justify-content: space-between; align-items: center; padding-bottom: 0.5rem; border-bottom: 1px solid var(--deck-border); font-size: 0.75rem; }
.pf-eco { padding: 1rem; border-radius: 16px; background: var(--deck-surface); border: 1px solid var(--deck-border); box-shadow: var(--pf-vet-shadow-md); }
.pf-eco__row { align-items: center; }
.pf-eco__flow { display: flex; flex-direction: column; align-items: center; gap: 0.35rem; text-align: center; }
.pf-eco__line { width: 100%; height: 2px; background: var(--deck-border); position: relative; }
.pf-eco__line::after {
  content: ''; position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%);
  width: 8px; height: 8px; border-radius: 999px; background: var(--deck-teal);
}
.pf-eco__sub {
  margin-top: 1rem; padding-top: 0.75rem; border-top: 1px solid var(--deck-border);
  display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.5rem; font-size: 0.75rem;
}
.pf-eco__sub-left { display: flex; align-items: center; gap: 0.65rem; }
.pf-phone {
  width: min(14rem, 100%); margin: 0 auto; padding: 0.75rem; border-radius: 24px;
  background: var(--deck-navy); color: #fff; border: 4px solid #374151; display: grid; gap: 0.65rem;
}
.pf-phone__bar { display: flex; justify-content: space-between; font-size: 0.6rem; font-family: var(--pf-font-mono, monospace); color: #9ca3af; }
.pf-phone__panel { padding: 0.6rem; border-radius: 8px; background: #1A2733; border: 1px solid #2D3D4E; display: grid; gap: 0.4rem; }
.pf-phone__event {
  display: flex; justify-content: space-between; align-items: center; gap: 0.35rem;
  padding: 0.4rem; border-radius: 6px; background: #0F1923; border-left: 2px solid var(--deck-teal); font-size: 0.65rem;
}
.pf-phone__event span.pf-muted { display: block; font-size: 0.55rem; color: #9ca3af; }
.pf-phone__cta {
  display: flex; align-items: center; justify-content: center; gap: 0.35rem;
  padding: 0.55rem; border-radius: 8px; background: var(--deck-teal); font-weight: 700; font-size: 0.75rem;
}
.pf-plan { display: flex; flex-direction: column; position: relative; }
.pf-plan__price { margin: 0.5rem 0; display: flex; align-items: baseline; gap: 0.35rem; }
.pf-plan__price strong { font-size: 1.5rem; font-family: var(--pf-font-mono, monospace); }
.pf-plan__note { margin-top: auto; padding-top: 0.5rem; border-top: 1px solid var(--deck-border); font-size: 0.7rem; }
.pf-plan--rec { border: 2px solid var(--deck-gold); }
.pf-plan--rec .pf-plan__note { border-top-color: rgba(255,255,255,0.15); }
.pf-plan--rec p { color: rgba(255,255,255,0.85); }
.pf-plan__rec {
  position: absolute; top: 0.5rem; right: 0.5rem; padding: 0.15rem 0.4rem; border-radius: 4px;
  background: var(--deck-gold); color: var(--deck-navy); font-size: 0.55rem; font-weight: 800; text-transform: uppercase; font-family: var(--pf-font-mono, monospace);
}
.pf-table-wrap { overflow: auto; border: 1px solid var(--deck-border); border-radius: 12px; }
.pf-table { width: 100%; border-collapse: collapse; font-size: 0.72rem; }
.pf-table th { background: var(--deck-navy); color: #fff; text-align: left; padding: 0.55rem 0.65rem; font-family: var(--pf-font-mono, monospace); font-size: 0.62rem; text-transform: uppercase; }
.pf-table td { padding: 0.55rem 0.65rem; border-top: 1px solid var(--deck-border); }
.pf-table__finance { background: color-mix(in srgb, var(--deck-bg) 70%, white); font-family: var(--pf-font-mono, monospace); }
.pf-demo { display: grid; gap: 0.75rem; }
.pf-demo__tabs { display: grid; grid-template-columns: repeat(5, 1fr); gap: 0.35rem; }
.pf-demo__tab {
  padding: 0.45rem 0.25rem; border-radius: 6px; border: 1px solid var(--deck-border);
  background: var(--deck-bg); color: var(--deck-navy); font-size: 0.62rem; font-family: var(--pf-font-mono, monospace); font-weight: 700; cursor: pointer;
}
.pf-demo__tab--active { background: var(--deck-navy); color: #fff; border-color: var(--deck-teal); }
.pf-demo__content { display: flex; gap: 0.85rem; align-items: flex-start; border-color: color-mix(in srgb, var(--deck-teal) 35%, white); }
.pf-demo__content h4 { margin: 0 0 0.35rem; font-size: 0.9rem; }
.pf-cta {
  margin: 0.5rem auto 0; padding: 0.7rem 1.4rem; border: none; border-radius: 12px;
  background: var(--deck-teal); color: #fff; font-weight: 700; font-size: 0.8rem; cursor: pointer;
  box-shadow: var(--pf-vet-shadow-md);
}
.pf-cta:hover { filter: brightness(0.95); transform: translateY(-1px); }
.pf-chip { display: inline-block; margin-top: 0.45rem; padding: 0.15rem 0.45rem; border-radius: 4px; background: rgba(255,255,255,0.1); color: var(--deck-gold); font-size: 0.62rem; font-family: var(--pf-font-mono, monospace); }
.pf-chip--sage { background: color-mix(in srgb, var(--deck-sage) 18%, white); color: var(--deck-sage); }
.pf-chip--teal { background: color-mix(in srgb, var(--deck-teal) 12%, white); color: var(--deck-teal); font-weight: 700; }
.pf-notes {
  position: absolute; bottom: 1rem; right: 1rem; width: min(22rem, calc(100% - 2rem));
  background: var(--deck-navy); color: #fff; border-radius: 12px; padding: 0.85rem;
  border: 1px solid color-mix(in srgb, var(--deck-teal) 40%, transparent); box-shadow: var(--pf-vet-shadow-hover); z-index: 30;
}
.pf-notes__head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; padding-bottom: 0.4rem; border-bottom: 1px solid #374151; font-size: 0.7rem; font-family: var(--pf-font-mono, monospace); font-weight: 700; color: var(--deck-gold); text-transform: uppercase; }
.pf-notes__head .pf-deck__icon-btn { color: #9ca3af; }
.pf-notes p { margin: 0; font-size: 0.75rem; line-height: 1.45; color: #e5e7eb; max-height: 10rem; overflow: auto; }
.pf-grid-modal {
  position: fixed; inset: 0; background: color-mix(in srgb, var(--deck-navy) 80%, transparent);
  backdrop-filter: blur(4px); z-index: 40; padding: 1.5rem; overflow: auto;
}
.pf-grid-modal__panel { max-width: 72rem; margin: 0 auto; }
.pf-grid-modal__head { display: flex; justify-content: space-between; align-items: flex-start; color: #fff; margin-bottom: 1rem; padding-bottom: 0.75rem; border-bottom: 1px solid #4b5563; }
.pf-grid-modal__head h3 { margin: 0; font-size: 1.15rem; }
.pf-grid-modal__head p { margin: 0.25rem 0 0; font-size: 0.75rem; color: #d1d5db; font-family: var(--pf-font-mono, monospace); }
.pf-grid-modal__head .pf-deck__icon-btn { color: #fff; background: rgba(255,255,255,0.1); }
.pf-grid-modal__list { display: grid; grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr)); gap: 0.75rem; }
.pf-grid-modal__item {
  text-align: left; padding: 0.75rem; border-radius: 12px; border: 1px solid #2D3D4E;
  background: #1A2733; color: #fff; cursor: pointer;
}
.pf-grid-modal__item--active, .pf-grid-modal__item:hover { border-color: var(--deck-teal); }
.pf-grid-modal__meta { display: flex; justify-content: space-between; font-size: 0.62rem; font-family: var(--pf-font-mono, monospace); color: #9ca3af; margin-bottom: 0.35rem; }
.pf-grid-modal__item strong { font-size: 0.75rem; line-height: 1.3; }
.pf-deck__nav { display: flex; gap: 0.65rem; }
.pf-btn {
  display: inline-flex; align-items: center; gap: 0.35rem; padding: 0.5rem 0.9rem; border-radius: 8px;
  font-size: 0.75rem; font-weight: 700; cursor: pointer; border: 1px solid var(--deck-border);
}
.pf-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.pf-btn--ghost { background: var(--deck-bg); color: var(--deck-navy); }
.pf-btn--primary { background: var(--deck-teal); color: #fff; border-color: transparent; box-shadow: var(--pf-vet-shadow-sm); }
.pf-deck__shortcuts { display: none; align-items: center; gap: 0.35rem; font-size: 0.7rem; font-family: var(--pf-font-mono, monospace); }
@media (min-width: 800px) { .pf-deck__shortcuts { display: flex; } }
kbd { padding: 0.1rem 0.35rem; border-radius: 4px; background: var(--deck-bg); border: 1px solid var(--deck-border); font-size: 0.62rem; }
.pf-teal, :deep(.pf-teal) { color: var(--deck-teal); }
.pf-gold, :deep(.pf-gold) { color: var(--deck-gold); }
.pf-sage, :deep(.pf-sage) { color: var(--deck-sage); }
.pf-alert, :deep(.pf-alert) { color: var(--deck-alert); }
.pf-muted { color: var(--deck-muted); }
.pf-muted-on-dark { color: #d1d5db; }
.pf-bold { font-weight: 700; }
.pf-mono { font-family: var(--pf-font-mono, monospace); }
.pf-small { font-size: 0.75rem; }
.pf-tiny { font-size: 0.65rem; }
.pf-center { text-align: center; }
.pf-left { text-align: left; }
.pf-row { display: inline-flex; align-items: center; gap: 0.35rem; }
</style>
