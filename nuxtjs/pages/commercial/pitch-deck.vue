<script setup lang="ts">
/** Ancien pitch-deck → présentation pages successives. */
definePageMeta({ middleware: [] })

await navigateTo(
  { path: '/presentation', query: { step: 'welcome' } },
  { redirectCode: 302, replace: true },
)
</script>
