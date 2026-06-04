<template>
  <view class="card">
    <text class="label">{{ label }}</text>
    <text class="value">{{ value }}</text>
  </view>
</template>

<script>
export default {
  props: {
    label: String,
    value: [String, Number]
  }
}
</script>

<style>
.card { padding: 16px; margin: 8px; border-radius: 8px; }
.label { color: #999; font-size: 12px; }
.value { color: #333; font-size: 18px; font-weight: bold; }
</style>
