import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Suspense, lazy } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';

const DashboardV2 = lazy(() => import('./pages/DashboardV2'));
const TestsPage = lazy(() => import('./pages/Tests'));
const Reports = lazy(() => import('./pages/Reports'));
const ReportDetailV2 = lazy(() => import('./pages/ReportDetailV2'));
const SettingsV2 = lazy(() => import('./pages/SettingsV2'));

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<Suspense fallback={<div/>}><DashboardV2 /></Suspense>} />
          <Route path="/tests" element={<Suspense fallback={<div/>}><TestsPage /></Suspense>} />
          <Route path="/reports" element={<Suspense fallback={<div/>}><Reports /></Suspense>} />
          <Route path="/reports/:testName" element={<Suspense fallback={<div/>}><ReportDetailV2 /></Suspense>} />
          <Route path="/settings" element={<Suspense fallback={<div/>}><SettingsV2 /></Suspense>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
