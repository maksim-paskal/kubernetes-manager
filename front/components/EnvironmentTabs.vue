<template>
  <div>
    <ul style="margin-top:10px" class="nav nav-tabs">
      <li class="nav-item" style="margin-left:20px">
        <b-link class="nav-link" active-class="active" to="/" exact-path>favorites</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="/my">created by me</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="/all">all</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="/remote-servers">remote servers</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="/wiki" exact-path>wiki</b-link>
      </li>
    </ul>
  </div>
</template>

<script>
export default {}
</script>