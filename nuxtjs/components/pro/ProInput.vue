<template>
  <div class="pro-field">
    <label v-if="label" :for="inputId" class="pro-label">{{ label }}</label>
    <input
      :id="inputId"
      :value="modelValue"
      :type="type"
      :name="name"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :required="required"
      :disabled="disabled"
      class="pro-input"
      :class="{ 'pro-input--error': !!error }"
      :data-testid="testId"
      @input="onInput"
    />
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
    error?: string
    testId?: string
  }>(),
  {
    type: 'text',
    required: false,
    disabled: false,
  },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const id = useId()
const inputId = computed(() => props.name || `pro-input-${id}`)

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>
