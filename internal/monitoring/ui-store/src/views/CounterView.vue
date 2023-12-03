<script setup lang="ts">
import {useCounterStore} from "../stores/counterStore";


const counterStore = useCounterStore();

const incrementCount = () => {
  counterStore.incrementCount();
};
</script>

<template>
  <div class="container">
    <div class="hero">
      <h1>Vue 3 Counter</h1>
      <p>This is a simple counter component that uses Bootstrap 5 toasts to display messages.</p>
    </div>

    <div class="counter">
      <button type="button" class="btn btn-primary" @click="incrementCount">Increment Count
        <strong>{{ counterStore.count }}</strong></button>
    </div>
  </div>
</template>

<style scoped>

</style>

<!--<div id="toast" class="toast hide" role="alert" aria-live="assertive" aria-atomic="true">-->
<!--<div class="toast-header">-->
<!--  <strong class="mr-auto">Vue 3 Counter</strong>-->
<!--  <small>Just now</small>-->
<!--  <button type="button" class="ml-2 mb-1 close" data-bs-dismiss="toast">-->
<!--    <span aria-hidden="true">&times;</span>-->
<!--  </button>-->
<!--</div>-->
<!--<div class="toast-body">-->
<!--  {{ toastMessage }}-->
<!--</div>-->
<!--</div>-->