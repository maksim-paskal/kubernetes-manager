<template>
  <b-dropdown :id="this.id" :text="this.buttonText" :variant="buttonVariant"
    toggle-class="text-muted text-start form-select dropdown-toggle-no-caret"
    style="width:100%;border: 1px solid #ced4da;" menu-class="dropdown-menu-top">
    <div style="padding:10px; ">
      <b-dropdown-form>
        <b-form-input :disabled="!this.isLoaded" style="width:100%;margin-bottom:10px" autocomplete="off"
          v-model="search" placeholder="Search">
        </b-form-input>
      </b-dropdown-form>
      <div style="height:224px;overflow:auto;">
        <b-spinner v-if="!this.isLoaded" variant="primary" />
        <b-alert v-else-if="this.errorText" variant="danger" show>{{ this.errorText }}</b-alert>
        <div v-else>
          <b-dropdown-item @click="select(item)" :active="selected == item ? true : false"
            v-for="(item, index) in this.filterResults()" :key="index">{{ item }}</b-dropdown-item>
        </div>
      </div>
      <b-button :disabled="!this.isLoaded" @click="reload()" variant="link" size="sm"
        class="text-decoration-none text-black"><em class="bi bi bi-arrow-repeat" /></b-button>
    </div>
  </b-dropdown>
</template>

<script>
export default {
  props: ['id', 'text', 'default', 'endpoint', 'value'],
  watch: {
    value: function () {
      this.selected = this.value;
    },
    search: function () {
      if (this.default && this.search == "") {
        this.reload();
      }
    }
  },
  mounted() {
    this.$root.$on('bv::dropdown::show', bvEvent => {
      if (bvEvent.componentId === this.id && !this.isLoaded) {
        this.load();
      }
    })

    // use custom value
    if (this.value) {
      this.selected = this.value;
    }

    // load default search value
    if (this.default) {
      this.search = this.default;
      // don't load results if default is set
      this.data = [this.default];
      this.isLoaded = true;
    }
  },
  data() {
    return {
      selected: "",
      search: "",
      errorText: null,
      isLoaded: false,
      data: [],
    }
  },
  computed: {
    buttonText() {
      if (this.selected) {
        return this.selected;
      } else {
        return this.text;
      }
    },
    buttonVariant() {
      if (this.selected == "") {
        return 'white ';
      } else {
        return 'custom';
      }
    }
  },
  methods: {
    reload() {
      this.isLoaded = false;
      this.load();
    },
    filterResults() {
      if (!this.isLoaded) return [];

      return this.data.filter(item => item.includes(this.search));
    },
    async load() {
      try {
        const result = await fetch(this.endpoint);
        if (result.ok) {
          const data = await result.json();
          this.data = data.Result
        } else {
          throw Error(await result.text());
        }
      } catch (e) {
        this.errorText = e;
      } finally {
        this.isLoaded = true
      }
    },
    select(data) {
      if (data == this.selected) {
        this.selected = "";
      } else {
        this.selected = data;
      }
      const item = { id: this.id, selected: this.selected }

      this.$store.commit("setDropDown", item)
      this.$nuxt.$emit('component::DropDown::selected', item)
    },
  }
}
</script>