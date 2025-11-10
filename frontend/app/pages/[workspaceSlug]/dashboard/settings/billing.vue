<template>
  <div class="space-y-10 w-full">
    <BillingSuccessMessage
      :show-success-message="showSuccessMessage"
      @close="showSuccessMessage = false"
    />

    <BillingCurrentPlan
      v-if="hasActiveSubscription"
      :user="user"
      :current-plan="currentPlan"
      :subscription-status="subscriptionStatus"
      :subscription-period-end="subscriptionPeriodEnd"
      :is-loading-portal="isLoadingPortal"
      :format-price="formatPrice"
      @manage-subscription="manageSubscription"
    />

    <BillingPlans
      v-else
      :products="products"
      :products-loading="productsLoading"
      :loading-product="loadingProduct"
      :recommended-plan-id="recommendedPlanId"
      :format-price="formatPrice"
      @subscribe="subscribeToPlan"
    />

    <BillingError :error="error" @close="error = null" />
  </div>
</template>

<script lang="ts" setup>
const route = useRoute()
const { user } = useAuth()
const { pb, refreshUser } = usePocketbase()
const { products, setProducts, formatPrice, formatInterval } = useProducts()
const { $api } = useNuxtApp()
const loadingProduct = ref<string | null>(null)
const isLoadingPortal = ref(false)
const error = ref<string | null>(null)
const showSuccessMessage = ref(false)

// Fetch products using useAsyncData with PocketBase SDK
const { data: fetchedProducts, pending: productsLoading } = await useAsyncData(
  'polar-products',
  async () => {
    const records = await pb.collection('polar_products').getFullList({
      filter: 'is_archived != true',
      sort: 'created'
    })

    // Transform records to match PolarProduct interface
    return records.map(record => ({
      id: record.id,
      name: record.name,
      description: record.description,
      price_amount: record.price_amount || 0,
      price_currency: record.price_currency || 'usd',
      recurring_interval: record.recurring_interval || 'month',
      recurring_interval_count: record.recurring_interval_count || 1,
      is_recurring: record.is_recurring || false,
      trial_interval: record.trial_interval,
      trial_interval_count: record.trial_interval_count,
      polar_price_id: record.polar_price_id || '',
      features: record.features || []
    }))
  }
)

setProducts(fetchedProducts.value as PolarProduct[])

// Check if redirected from successful checkout
onMounted(async () => {
  // Always refresh user data to get latest subscription status
  try {
    await refreshUser()
  } catch (error) {
    console.error('Failed to refresh user data:', error)
  }

  if (route.query.checkout === 'success') {
    showSuccessMessage.value = true
    // Remove query parameter from URL
    const { checkout, ...restQuery } = route.query
    navigateTo({ query: restQuery }, { replace: true })
  }
})

const subscriptionPeriodEnd = computed(() => {
  if (!user.value?.subscription_current_period_end) return ''
  return useDateFormat(user.value.subscription_current_period_end, 'DD/MM/YYYY')
    .value
})

const subscriptionStatus = computed(() => user.value?.subscription_status || '')

const hasActiveSubscription = computed(() => Boolean(subscriptionStatus.value))

const currentPlan = computed(() => {
  if (!user.value?.subscription_product_id) return null
  return products.value.find(
    (p: PolarProduct) => p.id === user.value?.subscription_product_id
  )
})

const recommendedPlanId = computed(
  () => products.value[1]?.id || products.value[0]?.id || null
)

const subscribeToPlan = async (planId: string) => {
  if (loadingProduct.value) return

  try {
    error.value = null
    loadingProduct.value = planId

    const workspaceSlug = route.params.workspaceSlug as string

    const response = await $api<{ url: string }>('api/checkout', {
      method: 'POST',
      body: {
        products: [planId],
        workspace_slug: workspaceSlug,
        return_path: '/dashboard/settings/billing'
      }
    })

    // Redirect to Polar checkout page
    if (response.url) {
      window.location.href = response.url
    } else {
      throw new Error('No checkout URL received')
    }
  } catch (err: any) {
    console.error('Error creating checkout session:', err)
    error.value =
      err.data?.error || err.message || 'Failed to create checkout session'
    loadingProduct.value = null
  }
}

const manageSubscription = async () => {
  if (isLoadingPortal.value) return

  try {
    error.value = null
    isLoadingPortal.value = true
    const workspaceSlug = route.params.workspaceSlug as string

    const response = await $api<{ url: string }>('api/customer-portal', {
      method: 'POST',
      body: {
        workspace_slug: workspaceSlug,
        return_path: '/dashboard/settings/billing'
      }
    })

    // Redirect to Polar customer portal
    if (response.url) {
      window.location.href = response.url
    } else {
      throw new Error('No portal URL received')
    }
  } catch (err: any) {
    console.error('Error creating customer portal session:', err)
    error.value =
      err.data?.error ||
      err.message ||
      'Failed to create customer portal session'
    isLoadingPortal.value = false
  }
}
</script>
