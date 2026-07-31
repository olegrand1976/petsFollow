/** Legacy `/ordonnances*` bookmarks → `/prescriptions*`. */
export default defineEventHandler((event) => {
  const path = getRequestURL(event).pathname
  if (path !== '/ordonnances' && !path.startsWith('/ordonnances/')) return
  return sendRedirect(event, path.replace(/^\/ordonnances/, '/prescriptions'), 301)
})
