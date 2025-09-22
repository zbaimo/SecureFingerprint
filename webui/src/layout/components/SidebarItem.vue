<template>
  <div v-if="!item.meta || !item.meta.hidden">
    <!-- 有子菜单的情况 -->
    <el-sub-menu
      v-if="hasChildren"
      :index="resolvePath(item.path)"
    >
      <template #title>
        <el-icon v-if="item.meta && item.meta.icon">
          <component :is="item.meta.icon" />
        </el-icon>
        <span>{{ item.meta && item.meta.title }}</span>
      </template>
      
      <sidebar-item
        v-for="child in visibleChildren"
        :key="child.path"
        :item="child"
        :base-path="resolvePath(child.path)"
      />
    </el-sub-menu>

    <!-- 单个菜单项 -->
    <el-menu-item
      v-else
      :index="resolvePath(item.path)"
    >
      <el-icon v-if="item.meta && item.meta.icon">
        <component :is="item.meta.icon" />
      </el-icon>
      <template #title>
        <span>{{ item.meta && item.meta.title }}</span>
      </template>
    </el-menu-item>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'SidebarItem',
  props: {
    item: {
      type: Object,
      required: true
    },
    basePath: {
      type: String,
      default: ''
    }
  },
  setup(props) {
    // 过滤可见的子菜单
    const visibleChildren = computed(() => {
      if (!props.item.children) return []
      return props.item.children.filter(child => !child.meta || !child.meta.hidden)
    })

    // 是否有子菜单
    const hasChildren = computed(() => {
      return visibleChildren.value.length > 0
    })

    // 解析路径
    const resolvePath = (routePath) => {
      if (routePath.startsWith('/')) {
        return routePath
      }
      return `${props.basePath}/${routePath}`.replace(/\/+/g, '/')
    }

    return {
      visibleChildren,
      hasChildren,
      resolvePath
    }
  }
}
</script>
