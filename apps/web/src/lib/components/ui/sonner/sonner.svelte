<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { Toaster as Sonner, type ToasterProps } from 'svelte-sonner';

  let {
    position = 'bottom-right',
    closeButton = true,
    richColors = false,
    duration = 4200,
    visibleToasts = 4,
    containerAriaLabel = t('web.sonner.stuffStashNotifications'),
    closeButtonAriaLabel = t('web.sonner.dismissNotification'),
    toastOptions = {},
    ...restProps
  }: ToasterProps = $props();
</script>

<Sonner
  {position}
  {closeButton}
  {richColors}
  {duration}
  {visibleToasts}
  {containerAriaLabel}
  {closeButtonAriaLabel}
  toastOptions={{
    ...toastOptions,
    classes: {
      toast: 'stuffstash-toast',
      title: 'stuffstash-toast-title',
      description: 'stuffstash-toast-description',
      actionButton: 'stuffstash-toast-action',
      cancelButton: 'stuffstash-toast-cancel',
      closeButton: 'stuffstash-toast-close',
      ...(toastOptions.classes ?? {})
    }
  }}
  {...restProps}
/>
