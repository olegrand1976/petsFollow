<template>
  <ProCard class="pro-mb-lg" data-testid="stock-deposits">
    <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.depositsTitle') }}</h3>
    <p class="pro-hint pro-mb-sm">{{ $t('pharmacy.stock.depositsHint') }}</p>
    <ul v-if="deposits.length" class="stock-deposits__list" data-testid="stock-deposits-list">
      <li v-for="d in deposits" :key="d.id" :data-testid="`stock-deposit-${d.id}`">
        <strong>{{ d.name }}</strong>
        <span class="pro-hint"> · {{ d.code }}</span>
        <ProBadge v-if="d.isDefault" variant="success">{{ $t('pharmacy.stock.depositDefault') }}</ProBadge>
      </li>
    </ul>
    <div v-if="canWrite" class="stock-form" data-testid="stock-deposit-create">
      <div>
        <label class="pro-label" for="deposit-name">{{ $t('pharmacy.stock.depositName') }}</label>
        <input id="deposit-name" v-model="form.name" class="pro-input" data-testid="stock-deposit-name">
      </div>
      <div>
        <label class="pro-label" for="deposit-code">{{ $t('pharmacy.stock.depositCode') }}</label>
        <input id="deposit-code" v-model="form.code" class="pro-input" data-testid="stock-deposit-code">
      </div>
      <label class="stock-deposits__check">
        <input v-model="form.isDefault" type="checkbox" data-testid="stock-deposit-default">
        {{ $t('pharmacy.stock.depositSetDefault') }}
      </label>
      <ProButton
        variant="secondary"
        test-id="stock-deposit-submit"
        :disabled="busy || !canCreate"
        @click="$emit('create')"
      >
        {{ $t('pharmacy.stock.depositCreate') }}
      </ProButton>
    </div>
    <p v-if="msg" class="pro-hint" data-testid="stock-deposit-msg">{{ msg }}</p>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacyDeposit, PharmacyDepositForm } from '~/composables/usePharmacyStockPage'

defineProps<{
  deposits: PharmacyDeposit[]
  canWrite: boolean
  busy: boolean
  msg: string
}>()
defineEmits<{ create: [] }>()

const form = defineModel<PharmacyDepositForm>('form', { required: true })

const canCreate = computed(() => Boolean(form.value.name.trim() && form.value.code.trim()))
</script>

<style scoped>
.stock-deposits__list {
  list-style: none;
  margin: 0 0 0.75rem;
  padding: 0;
  display: grid;
  gap: 0.35rem;
}
.stock-deposits__list li {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
}
.stock-form {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
  align-items: end;
}
.stock-deposits__check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.95rem;
}
</style>
