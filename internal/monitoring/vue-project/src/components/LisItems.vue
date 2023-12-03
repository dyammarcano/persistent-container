<script setup lang="ts">
import {defineComponent, onMounted, ref} from 'vue';
import axios from 'axios';

const ListItems = defineComponent({
  async setup() {
    // Initialize items as an empty array
    const items = ref([]);

    // Call the API endpoint when the component is created
    onMounted(async () => {
      try {
        const response = await axios.get('https://api.myapp.com/api/v1/users');

        // Set the local items data property as the list of usernames
        items.value = response.data.map(user => user.username);
      } catch (error) {
        console.error(`An error occurred while fetching the user data: ${error}`);
      }
    });

    return { items };
  },
  render() {
    return (
        <ul>
            {this.items.map((item) => (
                  <li>{item}</li>
              ))}
        </ul>
    );
  },
});
</script>

<template>
  <div>
    <a href="https://vitejs.dev" target="_blank">
      <img src="/vite.svg" class="logo" alt="Vite logo" />
    </a>
    <a href="https://vuejs.org/" target="_blank">
      <img src="./assets/vue.svg" class="logo vue" alt="Vue logo" />
    </a>
    <!-- Use ListItems component here -->
    <ListItems />
  </div>
</template>
