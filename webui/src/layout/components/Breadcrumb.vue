<template>
  <el-breadcrumb class="breadcrumb" separator="/">
    <el-breadcrumb-item
      v-for="(item, index) in breadcrumbs"
      :key="item.path"
    >
      <span
        v-if="item.redirect === 'noRedirect' || index === breadcrumbs.length - 1"
        class="no-redirect"
      >
        {{ item.meta.title }}
      </span>
      <a v-else @click.prevent="handleLink(item)">
        {{ item.meta.title }}
      </a>
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export default {
  name: 'Breadcrumb',
  setup() {
    const route = useRoute()
    const router = useRouter()
    const breadcrumbs = ref([])

    const getBreadcrumb = () => {
      // 过滤掉一些不需要显示在面包屑中的路由
      let matched = route.matched.filter(item => item.meta && item.meta.title)
      
      // 如果不是首页，添加首页到面包屑
      const first = matched[0]
      if (!isDashboard(first)) {
        matched = [{ 
          path: '/dashboard', 
          meta: { title: '首页' } 
        }].concat(matched)
      }

      breadcrumbs.value = matched.filter((item, index, arr) => {
        // 过滤重复项
        return item.meta && item.meta.title && item.meta.breadcrumb !== false
      })
    }

    const isDashboard = (route) => {
      const name = route && route.name
      if (!name) {
        return false
      }
      return name.trim().toLocaleLowerCase() === 'Dashboard'.toLocaleLowerCase()
    }

    const handleLink = (item) => {
      const { redirect, path } = item
      if (redirect) {
        router.push(redirect)
        return
      }
      router.push(path)
    }

    watch(route, getBreadcrumb, { immediate: true })

    return {
      breadcrumbs,
      handleLink
    }
  }
}
</script>

<style scoped>
.breadcrumb {
  display: inline-block;
  font-size: 14px;
  line-height: 50px;
  margin-left: 8px;
}

.breadcrumb .no-redirect {
  color: #97a8be;
  cursor: text;
}
</style>
