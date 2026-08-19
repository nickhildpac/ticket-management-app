import { createRootRouteWithContext, createRoute, createRouter, Outlet, redirect } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { Dashboard } from "@/features/dashboard";
import { Login } from "@/features/auth/login";
import { Signup } from "@/features/auth/signup";
import { AuthCallback } from "@/features/auth/callback";
import { TicketList } from "@/features/tickets/list";
import { AllTicketsList } from "@/features/tickets/all-tickets";
import { AssignedTicketsList } from "@/features/tickets/assigned-tickets";
import { TicketDetails } from "@/features/tickets/details";
import { TicketForm } from "@/features/tickets/form";
import { AdminPanel } from "@/features/admin";
import { DocumentUpload } from "@/features/admin/document-upload";
import { Profile } from "@/features/profile/profile";
import { getAuthUser, isAuthenticated, REDIRECT_PATH } from "@/app/auth";
import type { Role } from "@/lib/types";
import { validateTicketQueueSearch } from "@/features/tickets/queue-state";

export type AppContext = { qc: QueryClient; };

const rootRoute = createRootRouteWithContext<AppContext>()({
    component: () => <Outlet />,
});

// Auth Guard Helpers
const requireAuth = () => {
    if (!isAuthenticated()) {
        throw redirect({ to: '/login' });
    }
};

const getAuthUserRole = (): Role | null => {
    const role = getAuthUser()?.role;
    return role ? (role as Role) : null;
};

const requireAdmin = () => {
    if (!isAuthenticated()) {
        throw redirect({ to: '/login' });
    }
    const role = getAuthUserRole();
    if (!role) {
        throw redirect({ to: '/login' });
    }
    if (role !== 'admin') {
        throw redirect({ to: '/' });
    }
};

const requireAgent = () => {
    if (!isAuthenticated()) {
        throw redirect({ to: '/login' });
    }
    const role = getAuthUserRole();
    if (!role) {
        throw redirect({ to: '/login' });
    }
    if (role !== 'agent' && role !== 'admin') {
        throw redirect({ to: '/' });
    }
};

const requireDashboardAccess = () => {
    requireAuth();
    const role = getAuthUserRole();
    if (!role) {
        throw redirect({ to: '/login' });
    }
    if (role === 'user') {
        throw redirect({ to: '/tickets' });
    }
};

const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: Dashboard,
    beforeLoad: requireDashboardAccess
});

const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/login',
    component: Login,
});

const signupRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/signup',
    component: Signup,
});

// Where Keycloak redirects back to after authentication. Must stay in step with
// the client's registered redirect URIs in infra/keycloak/realm-export.json.
const authCallbackRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: REDIRECT_PATH,
    component: AuthCallback,
});

const ticketsListRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets',
    component: TicketList,
    validateSearch: validateTicketQueueSearch,
    beforeLoad: requireAuth
});

const ticketsAllRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/all',
    component: AllTicketsList,
    validateSearch: validateTicketQueueSearch,
    beforeLoad: requireAdmin
});

const ticketsAssignedRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/assigned',
    component: AssignedTicketsList,
    validateSearch: validateTicketQueueSearch,
    beforeLoad: requireAgent
});

const ticketCreateRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/new',
    component: TicketForm,
    beforeLoad: requireAuth // Should check role too
});

const ticketDetailsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/$id',
    component: TicketDetails,
    beforeLoad: requireAuth
});

const adminRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin',
    component: AdminPanel,
    beforeLoad: requireAdmin
});

const adminDocumentsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin/documents',
    component: DocumentUpload,
    beforeLoad: requireAdmin
});

const profileRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/profile',
    component: Profile,
    beforeLoad: requireAuth
});

const routeTree = rootRoute.addChildren([
    indexRoute,
    loginRoute,
    signupRoute,
    authCallbackRoute,
    ticketsListRoute,
    ticketsAllRoute,
    ticketsAssignedRoute,
    ticketCreateRoute,
    ticketDetailsRoute,
    adminRoute,
    adminDocumentsRoute,
    profileRoute,
]);

export const router = createRouter({
    routeTree,
    context: { qc: new QueryClient() } // Placeholder qc, real one passed in main
});

declare module '@tanstack/react-router' {
    interface Register {
        router: typeof router
    }
}
