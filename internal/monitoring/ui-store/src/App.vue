<script setup lang="ts">
import {useClockStore} from '@/stores/clockStore'
import {onMounted, onUnmounted} from 'vue'

const clockStore = useClockStore()
let intervalId: number

onMounted(() => {
  intervalId = <any> setInterval(clockStore.updateTime, 1000)
})

onUnmounted(() => {
  clearInterval(intervalId)
})

</script>

<template>
  <div>
    <header>
      <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
        <div class="container-fluid">
          <!-- your brand/logo link -->
          <router-link class="navbar-brand" to="#">
            <img src="/vite.svg" alt="logo" width="30" height="24" class="d-inline-block align-text-top">
          </router-link>

          <!-- button for small screens -->
          <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav"
                  aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
            <span class="navbar-toggler-icon"></span>
          </button>

          <!-- navbar content -->
          <div class="collapse navbar-collapse" id="navbarNav">

            <!-- main links  -->
            <ul class="navbar-nav me-auto mb-2 mb-lg-0">
              <li class="nav-item">
                <router-link class="nav-link" to="/">Home</router-link>
              </li>
              <li class="nav-item">
                <router-link class="nav-link" to="/list">List Data</router-link>
              </li>
              <li class="nav-item">
                <router-link class="nav-link" to="/add">Add Data</router-link>
              </li>
              <li class="nav-item">
                <router-link class="nav-link" to="/counter">Counter</router-link>
              </li>
            </ul>
          </div>
        </div>
      </nav>
    </header>
    <main class="flex-shrink-0">
      <router-view></router-view>
    </main>
<!--    <footer class="footer mt-auto py-3 bg-light">-->
    <!--      <div class="container text-center">-->
    <!--        <span class="text-muted">© 2023 <a href=https://www.wikiwand.com/pt/ACME target="_blank">ACME</a> Corporation</span>-->
    <!--      </div>-->
    <!--    </footer>-->
  </div>
</template>

<style scoped>

</style>
