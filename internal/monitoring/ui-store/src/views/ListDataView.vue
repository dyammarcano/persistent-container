<script setup lang="ts">
import {onMounted, ref} from 'vue';
import axios from 'axios';

interface Item {
  id: number;
  title: string;
}

const items = ref<Item[]>([]);
const itemToDelete = ref<Item | null>(null);

axios.interceptors.request.use((config) => {
  const token = '3B3sZjQcU2Pr8stjtBsKDtZyocA9D3DRHNbyf1y3uaUCxoEmYcBM9dph2QqLtS4YsNbbpAYik9AiU8XEVJ5Au8pVvW34nH1SjTU3XLWweJ681VVCXz65WzhZVHCEarRTm9G7mSDuQoAUiNZjdw1g9FnjW8LWeSigMKRGRDdfJgQktx8iS5WuNh15yWL7jjLXa9W3Zzj84Z9hmM31E2Fu2Prtnwx2tUZfC5CwVzzZV4ZdGSizMbUGnTMzBWh2dsT7TmzFvJZm86AD4rNhoyw7guPn4Lmm41hCb8odYr1FktqLWrkAnLAQrNh3tP9dEBoh';

  config.headers.Authorization = `Bearer ${token}`;

  return config;
}, (error) => {
  return Promise.reject(error);
});

onMounted(async () => {
  const {data} = await axios.get('http://localhost:8080/api/v1/data');
  items.value = data;
});

const prepareDelete = ({item}: { item: any }) => {
  itemToDelete.value = item;
}

const deleteItem = async ({itemToDelete}: { itemToDelete: any }) => {
  await axios.delete(`http://localhost:8080/api/v1/data/${itemToDelete.id}`);
  items.value = items.value.filter(item => {
    return item.id !== itemToDelete.id;
  });
  itemToDelete.value = null;
}

const viewItem = ({id}: { id: any }) => {
  console.log(`Viewing item ${id}`)
  // implement logic to view specific item details
}

</script>

<template>
  <div class="container">
    <div class="hero">
      <h1>List Data</h1>
    </div>

    <ol class="list-group list-group-numbered">
      <li class="list-group-item d-flex justify-content-between align-items-start" v-for="item in items" :key="item.id">
        {{ item.title }}

        <div class="btn-group" role="group" aria-label="Basic example">
          <button type="button" class="btn btn-secondary" @click="viewItem({id : item.id})">View</button>
          <button type="button" class="btn btn-danger" @click="prepareDelete({item})" data-bs-toggle="modal"
                  data-bs-target="#deleteModal">Delete
          </button>
        </div>

      </li>
    </ol>

    <!-- Bootstrap modal for deletion confirmation -->
    <div class="modal" id="deleteModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Delete Confirmation</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <p>Are you sure you want to delete this item?</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-danger" @click="deleteItem({itemToDelete : {item : itemToDelete}})"
                    data-bs-dismiss="modal">Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>