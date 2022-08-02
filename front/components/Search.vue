<template>
  <div class="dropdown">
    <b-form-input autocomplete="off" v-model="searchText" style="width:500px" placeholder="Type to Search"
      id="dropdownMenuButton1" data-bs-toggle="dropdown" aria-expanded="false" />
    <ul class="dropdown-menu" style="width:500px" aria-labelledby="dropdownMenuButton1">
      <div v-if="loading" style="padding:10px">
        <b-spinner variant="primary" />
      </div>
      <div v-else-if="searchText.length == 0" style="padding:10px">start typing</div>
      <div v-else-if="searchedData.length == 0" style="padding:10px">no result</div>
      <li v-for="(item, index) in searchedData" :key="index">
        <b-link @click="navigate(item)" class="dropdown-item">{{ getEnvironmentName(item) }}</b-link>
      </li>
    </ul>
  </div>
</template>

<script>
import bootstrap from 'bootstrap/dist/js/bootstrap.bundle.js';

export default {
  data() {
    return {
      loading: false,
      searchText: '',
      dataLoaded: false,
      data: [],
      searchedData: []
    }
  },
  mounted() {
    let dropdownElementList = [].slice.call(document.querySelectorAll('.dropdown-toggle'))
    dropdownElementList.map(function (dropdownToggleEl) {
      return new bootstrap.Dropdown(dropdownToggleEl)
    })
  },
  watch: {
    searchText() {
      if (!this.dataLoaded) {
        this.load()
      }

      if (this.searchText.length > 0) {
        this.search()
      } else {
        this.searchedData = []
      }
    }
  },
  methods: {
    navigate(data) {
      this.loadEnvironment(data.ID)
      this.$router.push(`/${data.ID}/info`)
    },
    getEnvironmentName(data) {
      if (data.NamespaceAnnotations && data.NamespaceAnnotations[this.const().LabelEnvironmentName]) {
        return data.NamespaceAnnotations[this.const().LabelEnvironmentName]
      }

      return data.Namespace
    },
    search() {
      if (!this.dataLoaded) return;

      this.searchedData = []

      this.data.forEach(el => {
        if (JSON.stringify(el).toLowerCase().includes(this.searchText.toLowerCase())) {
          this.searchedData.push(el)
          if (this.searchedData.length >= 10) return;
        }
      });

    },
    async load() {
      if (this.loading) return;

      this.loading = true
      try {
        const result = await fetch('/api/environments')
        if (result.ok) {
          const data = await result.json();
          this.data = data.Result
        }
      } finally {
        this.loading = false
        this.dataLoaded = true
        this.search()
      }
    }
  }
}
</script>