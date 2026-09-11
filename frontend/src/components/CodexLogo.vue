<script setup lang="ts">
import appIcon from '@/assets/app-icon.svg';
defineProps<{size?:number;withDot?:boolean}>();
</script>
<template><img :src="appIcon" :width="size ?? 22" :height="size ?? 22" alt="" aria-hidden="true" /></template>
