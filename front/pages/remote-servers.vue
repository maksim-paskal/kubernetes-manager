<template>
  <div>
    <b-spinner v-if="!this.user.user" variant="primary" />
    <RemoteList v-else />
  </div>
</template>