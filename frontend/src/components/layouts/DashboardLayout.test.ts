import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import DashboardLayout from '../layouts/DashboardLayout.vue'

interface SidebarItem {
  label: string
  path: string
  icon?: string
}

describe('DashboardLayout', () => {
  it('renders default sidebar items', () => {
    const wrapper = mount(DashboardLayout, {
      slots: {
        default: '<div>Content</div>',
      },
    })

    expect(wrapper.text()).toContain('Phosboard')
    expect(wrapper.text()).toContain('Dashboard')
    expect(wrapper.text()).toContain('Documents')
    expect(wrapper.text()).toContain('Sources')
    expect(wrapper.text()).toContain('Users')
    expect(wrapper.text()).toContain('Settings')
  })

  it('renders custom sidebar items', () => {
    const customItems: SidebarItem[] = [
      { label: 'Custom 1', path: '/custom1' },
      { label: 'Custom 2', path: '/custom2' },
    ]

    const wrapper = mount(DashboardLayout, {
      props: {
        sidebarItems: customItems,
      },
      slots: {
        default: '<div>Content</div>',
      },
    })

    expect(wrapper.text()).toContain('Custom 1')
    expect(wrapper.text()).toContain('Custom 2')
    expect(wrapper.text()).not.toContain('Dashboard')
  })

  it('renders default slot content', () => {
    const wrapper = mount(DashboardLayout, {
      slots: {
        default: '<p class="test-content">Test Content</p>',
      },
    })

    expect(wrapper.find('.test-content').exists()).toBe(true)
    expect(wrapper.find('.test-content').text()).toBe('Test Content')
  })

  it('has correct layout structure', () => {
    const wrapper = mount(DashboardLayout, {
      slots: {
        default: '<div>Content</div>',
      },
    })

    expect(wrapper.find('aside').exists()).toBe(true)
    expect(wrapper.find('main').exists()).toBe(true)
  })
})