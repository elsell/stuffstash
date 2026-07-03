<script lang="ts">
  import AssetDetailHero, { type DetailPhoto } from './AssetDetailHero.svelte';

  let {
    heroPhoto,
    photos,
    onSelectPhoto
  }: {
    heroPhoto: DetailPhoto;
    photos: DetailPhoto[];
    onSelectPhoto: (photoId: string) => void;
  } = $props();
</script>

<AssetDetailHero
  kind="item"
  {heroPhoto}
  {photos}
  canAddPhoto={true}
  uploadError=""
  onChoosePhoto={() => {}}
  {onSelectPhoto}
>
  <div class="asset-detail-copy">
    <h1>Identity</h1>
  </div>
</AssetDetailHero>
