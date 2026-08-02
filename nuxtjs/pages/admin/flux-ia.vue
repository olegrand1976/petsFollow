<script setup lang="ts">
/** Alias admin → `/flux-ia` en préservant la query (`?profile=`). */
definePageMeta({ middleware: [] })

const route = useRoute()
await navigateTo(
  { path: '/flux-ia', query: route.query },
  { redirectCode: 302, replace: true },
)
</script>
