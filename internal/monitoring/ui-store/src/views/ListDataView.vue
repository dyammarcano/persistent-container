<script setup lang="ts">
import {onMounted, ref} from 'vue';
import axios from 'axios';
import {Modal} from 'bootstrap';

interface Item {
  id: string;
}

const items = ref<Item[]>([]);

const setAuthorizationToken = () => {
  const token = '3B3sZjQcU2Pr8stjtBsKDtZyocA9D3DRHNbyf1y3uaUCxoEmYcBM9dph2QqLtS4YsNbbpAYik9AiU8XEVJ5Au8pVvW34nH1SjTU3XLWweJ681VVCXz65WzhZVHCEarRTm9G7mSDuQoAUiNZjdw1g9FnjW8LWeSigMKRGRDdfJgQktx8iS5WuNh15yWL7jjLXa9W3Zzj84Z9hmM31E2Fu2Prtnwx2tUZfC5CwVzzZV4ZdGSizMbUGnTMzBWh2dsT7TmzFvJZm86AD4rNhoyw7guPn4Lmm41hCb8odYr1FktqLWrkAnLAQrNh3tP9dEBoh';
  axios.interceptors.request.use((config) => {
    config.headers.Authorization = `Bearer ${token}`;
    return config;
  }, Promise.reject);
};

const fetchItems = async () => {
  items.value = await getItem();
};

const getItem = async () => {
  try {
    const response = await axios.get('/api/v1/data');
    return response.data;
  } catch (error) {
    console.error(error);
  }
};

const removeItem = async (id: string) => {
  await deleteItem(id);
  items.value = items.value.filter(item => item.id !== id);
};

let deleteModal;
let viewModal;

onMounted(() => {
  setAuthorizationToken();
  fetchItems();

  // initialize your modal
  deleteModal = new Modal(document.getElementById('deleteModal'));
  viewModal = new Modal(document.getElementById('viewModal'));
});

const prepareDelete = (id: string) => {
  itemToBeDeleted.value = id;

  // show the modal when delete button clicked
  deleteModal.show();
};

const itemToBeDeleted = ref<string>();
const itemToBeViewed = ref<string>();

const deleteItem = async (id: string) => {
  await axios.delete(`/api/v1/data/${id}`);
  items.value = items.value.filter(item => {
    return item.id !== id;
  });
}

const viewItem = async (id: string) => {
  try {
    const response = await axios.get(`/api/v1/data/${id}`);
    // console.log(response.data);
    itemToBeViewed.value = JSON.stringify(response.data, null, 2);
  } catch (error) {
    console.error(error);
  }

  // show the modal when view button clicked
  viewModal.show();
};

</script>

<template>
  <div class="container">
    <div class="hero">
      <h1>List Data</h1>
    </div>

    <div class="container-sm">
      <ol class="list-group list-group-numbered">
        <li class="list-group-item d-flex justify-content-between align-items-start" v-for="item in items"
            :key="item.id">

          <div class="ms-2 me-auto">
            <!--            <div class="fw-bold">{{ item.id }}</div>-->
            {{ item.id }}
          </div>

          <div class="btn-group" role="group" aria-label="Basic example">
            <button type="button" class="btn btn-secondary" @click="viewItem(item.id)">View</button>
            <button type="button" class="btn btn-danger" @click="prepareDelete(item.id)" data-bs-toggle="modal"
                    data-bs-target="#deleteModal">Delete
            </button>
          </div>
        </li>
      </ol>

      <!-- Bootstrap modal for deletion confirmation -->
      <div class="modal fade" id="deleteModal" tabindex="-1">
        <div class="modal-dialog modal-dialog-centered">
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
              <button type="button" class="btn btn-danger" @click="removeItem(<string>itemToBeDeleted)"
                      data-bs-dismiss="modal">
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- End of Bootstrap modal for deletion confirmation -->

    <!-- Bootstrap modal for view data -->
    <div class="modal fade" id="viewModal" tabindex="-1">
      <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">View Data</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            {{ itemToBeViewed }}
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container-sm {
  width: 60%;
}
</style>