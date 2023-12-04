<!--<script setup lang="ts">-->
<!--import router from "../router";-->
<!--import axios from "axios"; // adjust the import based on your project structure-->

<!--axios.interceptors.request.use((config) => {-->
<!--  const token = '3B3sZjQcU2Pr8stjtBsKDtZyocA9D3DRHNbyf1y3uaUCxoEmYcBM9dph2QqLtS4YsNbbpAYik9AiU8XEVJ5Au8pVvW34nH1SjTU3XLWweJ681VVCXz65WzhZVHCEarRTm9G7mSDuQoAUiNZjdw1g9FnjW8LWeSigMKRGRDdfJgQktx8iS5WuNh15yWL7jjLXa9W3Zzj84Z9hmM31E2Fu2Prtnwx2tUZfC5CwVzzZV4ZdGSizMbUGnTMzBWh2dsT7TmzFvJZm86AD4rNhoyw7guPn4Lmm41hCb8odYr1FktqLWrkAnLAQrNh3tP9dEBoh';-->

<!--  config.headers.Authorization = `Bearer ${token}`;-->

<!--  return config;-->
<!--}, (error) => {-->
<!--  return Promise.reject(error);-->
<!--});-->

<!--const viewItem = ({id}: { id: any }) => {-->
<!--  let url = document.location.origin + '/api/v1/data/' + id;-->
<!--  console.log(`Viewing item ${url}`);-->

<!--  router.push({name: 'ViewItem', params: {id: id}});-->
<!--}-->

<!--</script>-->

<!--<template>-->
<!--  <div class="hero">-->
<!--    <h1>View Item</h1>-->
<!--  </div>-->

<!--  <div class="container-sm">-->
<!--    {{ viewItem() }}-->
<!--  </div>-->
<!--</template>-->

<!--<style scoped>-->

<!--</style>-->