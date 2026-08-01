<script setup lang="ts">
/**
 * Alias commercial-manager → `/flux` en préservant la query (`?profile=`).
 */
definePageMeta({ middleware: [] })

const route = useRoute()
await navigateTo(
  { path: '/flux', query: route.query },
  { redirectCode: 302, replace: true },
)
</script>