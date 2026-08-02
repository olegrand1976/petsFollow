<script setup lang="ts">
/** Alias manager → `/presentation` en préservant la query (`?step=`). */
definePageMeta({ middleware: [] })

const route = useRoute()
await navigateTo(
  { path: '/presentation', query: route.query },
  { redirectCode: 302, replace: true },
)
</script>
