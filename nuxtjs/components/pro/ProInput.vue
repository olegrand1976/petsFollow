<template>
  <div class="pro-field">
    <label v-if="label" :for="inputId" class="pro-label">{{ label }}</label>
    <div class="pro-input-wrap" :class="{ 'pro-input-wrap--reveal': revealable }">
      <input
        :id="inputId"
        :value="modelValue"
        :type="effectiveType"
        :name="name"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :required="required"
        :disabled="disabled"
        :maxlength="maxlength"
        class="pro-input"
        :class="{ 'pro-input--error': !!error }"
        :data-testid="testId"
        @input="onInput"
      />
      <button
        v-if="revealable"
        type="button"
        class="pro-input-reveal"
        :aria-label="revealed ? hideLabel : showLabel"
        :aria-pressed="revealed"
        :data-testid="testId ? `${testId}-reveal` : undefined"
        :disabled="disabled"
        @click="revealed = !revealed"
      >
        <ProIcon :name="revealed ? 'visibility_off' : 'visibility'" :size="22" />
      </button>
    </div>
    <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue: string
    label?: string
    type?: string
    name?: string
    placeholder?: string
    autocomplete?: string
    required?: boolean
    disabled?: boolean
    maxlength?: number | string
    error?: string
    testId?: string
    /** Affiche un bouton œil pour basculer type password ↔ text. */
    revealable?: boolean
  }>(),
  {
    type: 'text',
    required: false,
    disabled: false,
    revealable: false,
  },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const { t } = useI18n()

const id = useId()
const inputId = computed(() => props.name || `pro-input-${id}`)
const revealed = ref(false)
const effectiveType = computed(() => {
  if (props.revealable && props.type === 'password') {
    return revealed.value ? 'text' : 'password'
  }
  return props.type
})
const showLabel = computed(() => t('auth.fields.showPassword'))
const hideLabel = computed(() => t('auth.fields.hidePassword'))

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>
