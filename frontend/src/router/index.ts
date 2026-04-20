import { createRouter, createWebHistory } from "vue-router";
import CompanyDetailPage from "../pages/CompanyDetailPage.vue";
import CompanyListPage from "../pages/CompanyListPage.vue";
import EditCompanyPage from "../pages/EditCompanyPage.vue";
import EditPostPage from "../pages/EditPostPage.vue";
import LoginPage from "../pages/LoginPage.vue";
import NewCompanyPage from "../pages/NewCompanyPage.vue";
import NewPostPage from "../pages/NewPostPage.vue";
import NotFoundPage from "../pages/NotFoundPage.vue";
import PostDetailPage from "../pages/PostDetailPage.vue";
import RootRedirectPage from "../pages/RootRedirectPage.vue";
import SignupPage from "../pages/SignupPage.vue";
import { useAuth } from "../composables/useAuth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: RootRedirectPage },
    { path: "/login", component: LoginPage, meta: { guestOnly: true } },
    { path: "/signup", component: SignupPage, meta: { guestOnly: true } },
    { path: "/companies", component: CompanyListPage },
    { path: "/companies/new", component: NewCompanyPage, meta: { requiresAuth: true } },
    { path: "/companies/:id", component: CompanyDetailPage },
    {
      path: "/companies/:id/posts/new",
      component: NewPostPage,
      meta: { requiresAuth: true },
    },
    {
      path: "/companies/:id/edit",
      component: EditCompanyPage,
      meta: { requiresAuth: true },
    },
    { path: "/posts/:postId", component: PostDetailPage },
    {
      path: "/posts/:postId/edit",
      component: EditPostPage,
      meta: { requiresAuth: true },
    },
    { path: "/:pathMatch(.*)*", component: NotFoundPage },
  ],
});

router.beforeEach(async (to) => {
  const auth = useAuth();
  await auth.ensureBootstrapped();

  if (to.meta.requiresAuth && !auth.isAuthenticated.value) {
    return {
      path: "/login",
      query: { redirect: to.fullPath },
    };
  }

  if (to.meta.guestOnly && auth.isAuthenticated.value) {
    return { path: "/companies" };
  }

  if (to.path === "/") {
    return { path: auth.isAuthenticated.value ? "/companies" : "/login" };
  }

  return true;
});

export default router;
