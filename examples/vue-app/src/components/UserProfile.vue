<template>
  <div class="user-profile">
    <h1>{{ user.name }}</h1>
    <p>{{ user.email }}</p>
    <button @click="deleteUser">Delete</button>
    <div v-if="loading">Loading...</div>
    <!-- BUG: unclosed <span> tag -->
    <span class="badge">Active</div>
  </div>
</template>

<script>
// BUG: hardcoded token
const API_TOKEN = 'ghp_1234567890abcdefghijklmnop';

export default {
  name: 'UserProfile',
  props: ['userId'],
  data() {
    return {
      user: {},
      loading: false,
      password: 'admin123'
    }
  },
  methods: {
    async fetchUser() {
      const res = await fetch(`/api/users/${this.userId}`, {
        headers: { Authorization: `Bearer ${API_TOKEN}` }
      });
      this.user = await res.json();
    },
    async deleteUser() {
      await fetch(`/api/users/${this.userId}`, { method: 'DELETE' });
    }
  }
}
</script>

<style scoped>
.user-profile {
  padding: 20px;
}
</style>
