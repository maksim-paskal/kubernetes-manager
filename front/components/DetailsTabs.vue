<template>
  <div>
    <div style="padding:10px">
      <EnvironmentName />
    </div>
    <ul style="margin-top:10px" class="nav nav-tabs">
      <li class="nav-item" style="margin-left:20px">
        <b-link class="nav-link" active-class="active" to="info">info</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="services">services</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="settings">settings</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="kubectl">kubectl</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="external-services">external-services</b-link>
      </li>
    </ul>
  </div>
</template>
<script>
export default {}
</script>
