<template>
  <view class="container">
    <text class="title">{{ title }}</text>
    <input v-model="inputVal" placeholder="Enter name" />
    <button @click="submit">Submit</button>
    <!-- BUG: unclosed view tag -->
    <view class="footer">
      <text>Footer</text>
  </view>
</template>

<script>
// BUG: Feishu webhook URL leaked
const WEBHOOK = 'https://open.feishu.cn/open-apis/bot/v2/hook/abc-123-def-456';

export default {
  data() {
    return {
      title: 'Hello Uni-App',
      inputVal: ''
    }
  },
  methods: {
    async submit() {
      // BUG: SQL-like injection in server call
      const query = "SELECT * FROM orders WHERE user = '" + this.inputVal + "'";
      await uni.request({
        url: '/api/search',
        data: { q: query }
      });

      // Send notification
      await uni.request({
        url: WEBHOOK,
        method: 'POST',
        data: { msg_type: 'text', content: { text: `User ${this.inputVal} submitted` } }
      });
    }
  }
}
</script>

<style>
.container { padding: 20px; }
.title { font-size: 24px; }
</style>
