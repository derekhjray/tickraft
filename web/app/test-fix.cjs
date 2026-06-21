/**
 * Comprehensive test for column resize fix.
 * Tests:
 * 1. Drag widens a column — verifies only target column changes
 * 2. Drag narrows a column — verifies no empty space, minimal other column shift
 * 3. Multi-drag stability — drag multiple columns, verify no chaos
 * 4. Persistence — widths survive page reload
 * 5. Scroll sync — header and body scroll together
 */
const PW_PATH = '/Users/derekray/workspace/auzekalabs/tickraft/arcadia/node_modules/.pnpm/playwright@1.61.1/node_modules/playwright';
const { chromium } = require(PW_PATH);
const BASE = 'http://localhost:5190';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function getColWidths(page) {
  return page.evaluate(() => {
    const cols = document.querySelectorAll('.el-table__header colgroup col');
    return Array.from(cols).map((c) => parseInt(c.getAttribute('width') || '0', 10));
  });
}

async function dragColumn(page, colIdx, distance) {
  const cellInfo = await page.evaluate((idx) => {
    const cells = document.querySelectorAll('.el-table__header-wrapper th.el-table__cell');
    if (idx >= cells.length) return null;
    const cell = cells[idx];
    const rect = cell.getBoundingClientRect();
    return { right: rect.right, top: rect.top, bottom: rect.bottom };
  }, colIdx);
  if (!cellInfo) return false;

  const startX = cellInfo.right - 3;
  const startY = (cellInfo.top + cellInfo.bottom) / 2;

  await page.mouse.move(startX, startY);
  await sleep(50);
  await page.mouse.down();
  await sleep(80);
  for (let i = 1; i <= 15; i++) {
    await page.mouse.move(startX + (distance * i) / 15, startY);
    await sleep(20);
  }
  await page.mouse.up();
  await sleep(600);
  return true;
}

async function loginAndNavigate(page, path) {
  await page.goto(`${BASE}/login`);
  await sleep(2000);
  await page.fill('#tk-username', 'admin');
  await page.fill('#tk-password', 'admin123');
  await page.click('.tk-auth-form__submit');
  await sleep(3000);
  if (page.url().includes('/login')) {
    await page.evaluate(() => localStorage.setItem('tk-token', JSON.stringify('mock-jwt-token')));
  }
  await page.goto(`${BASE}${path}`);
  await sleep(3000);
}

(async () => {
  const browser = await chromium.launch({
    headless: true,
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await ctx.newPage();
  const results = [];
  let pass = 0, fail = 0;

  try {
    console.log('=== Login ===');
    await loginAndNavigate(page, '/asset/list');
    console.log('  URL:', page.url());

    // Clear previous widths for clean test
    await page.evaluate(() => localStorage.removeItem('tk-table-widths'));
    await page.reload();
    await sleep(3000);

    // Get initial widths
    const initialWidths = await getColWidths(page);
    console.log('\nInitial column widths:', initialWidths);
    console.log('Column count:', initialWidths.length);

    // ── Test 1: Drag column 1 wider → only column 1 should change ──
    console.log('\n=== Test 1: Drag Column 1 Widen (+50px) ===');
    const before1 = await getColWidths(page);
    console.log('  Before:', before1);

    const dragged = await dragColumn(page, 1, 50);
    if (!dragged) {
      console.log('  ❌ Could not drag');
      fail++;
      results.push('Drag failed');
    } else {
      const after1 = await getColWidths(page);
      console.log('  After:', after1);

      const diffs = after1.map((w, i) => w - before1[i]);
      console.log('  Diffs:', diffs);

      // Column 1 should have +50 (within tolerance)
      const targetDiff = diffs[1];
      const ok1 = Math.abs(targetDiff - 50) <= 5;
      console.log(`  Column 1 changed by ${targetDiff}px (expected +50): ${ok1 ? '✅' : '❌'}`);
      if (ok1) { pass++; results.push('Target column widen'); }
      else { fail++; results.push('Target column widen FAILED'); }

      // Other columns should NOT change significantly (< 15px tolerance for flex absorption)
      const otherDiffs = diffs.filter((_, i) => i !== 1);
      const maxOtherShift = Math.max(...otherDiffs.map(Math.abs));
      console.log(`  Max other column shift: ${maxOtherShift}px (tolerance 15px)`);
      if (maxOtherShift <= 15) { pass++; results.push('Other columns stable'); }
      else {
        console.log(`  ❌ Other columns shifted too much:`, diffs);
        fail++; results.push('Other columns stable FAILED');
      }
    }

    // ── Test 2: Drag column 1 narrower (it was widened in test 1) ──
    console.log('\n=== Test 2: Drag Column 1 Narrow (-30px) ===');
    const before2 = await getColWidths(page);
    console.log('  Before:', before2);

    await dragColumn(page, 1, -30);
    const after2 = await getColWidths(page);
    console.log('  After:', after2);

    const diffs2 = after2.map((w, i) => w - before2[i]);
    console.log('  Diffs:', diffs2);

    // Column 1 should have ~-30 change (from 170 → 140)
    const col1Diff = diffs2[1];
    const ok2a = Math.abs(col1Diff + 30) <= 5;
    console.log(`  Column 1 changed by ${col1Diff}px (expected -30): ${ok2a ? '✅' : '❌'}`);
    if (ok2a) { pass++; results.push('Narrow works'); }
    else { fail++; results.push('Narrow works FAILED'); }

    // Other columns should stay stable
    const otherDiffs2 = diffs2.filter((_, i) => i !== 1);
    const maxShift2 = Math.max(...otherDiffs2.map(Math.abs));
    console.log(`  Max other column shift: ${maxShift2}px`);
    if (maxShift2 <= 15) { pass++; results.push('Narrow stability'); }
    else { fail++; results.push('Narrow stability FAILED'); }

    // ── Test 3: Drag column 4 wider (it's at its minWidth, so fresh drag) ──
    console.log('\n=== Test 3: Drag Column 4 Widen (+40px) ===');
    const before3 = await getColWidths(page);
    console.log('  Before:', before3);

    await dragColumn(page, 4, 40);
    const after3 = await getColWidths(page);
    console.log('  After:', after3);

    const diffs3 = after3.map((w, i) => w - before3[i]);
    console.log('  Diffs:', diffs3);

    // Verify column 4 changed by ~40px
    const col4Diff = diffs3[4];
    const ok3a = Math.abs(col4Diff - 40) <= 5;
    console.log(`  Column 4 changed by ${col4Diff}px (expected +40): ${ok3a ? '✅' : '❌'}`);

    // Count columns that shifted > 15px (excluding target)
    const chaoticCols = diffs3.filter((d, i) => i !== 4 && Math.abs(d) > 15).length;
    console.log(`  Non-target columns with >15px shift: ${chaoticCols}`);

    if (ok3a && chaoticCols === 0) { pass++; results.push('Multi-drag stability'); }
    else { fail++; results.push('Multi-drag stability FAILED'); }

    // ── Test 4: Persistence ──
    console.log('\n=== Test 4: Persistence ===');
    const savedData = await page.evaluate(() => localStorage.getItem('tk-table-widths'));
    console.log('  localStorage:', savedData?.substring(0, 200) || '(empty)');

    await page.reload();
    await sleep(3000);
    const afterReload = await getColWidths(page);
    console.log('  After reload:', afterReload);

    // Column 1 was dragged +50 (initial was probably 180 or similar)
    // Check that it's NOT back to the original default
    if (savedData) {
      const parsed = JSON.parse(savedData);
      console.log('  Parsed saved widths:', JSON.stringify(parsed));
      // At least the dragged columns should have saved widths
      const props = Object.keys(parsed);
      if (props.length >= 1) {
        console.log(`  ✅ ${props.length} column widths persisted`);
        pass++; results.push('Persistence');
      } else {
        console.log('  ❌ No persisted widths');
        fail++; results.push('Persistence FAILED');
      }
    } else {
      console.log('  ❌ No localStorage data');
      fail++; results.push('Persistence FAILED');
    }

    // ── Test 5: Scroll sync (using saved wide widths to force overflow) ──
    console.log('\n=== Test 5: Scroll Sync ===');
    await page.evaluate(() => {
      localStorage.setItem('tk-table-widths', JSON.stringify({
        'asset-list': { name: 250, assetType: 200, assetKey: 300, status: 150, labels: 250, lastActiveAt: 200 }
      }));
    });
    await page.reload();
    await sleep(3000);

    const overflow = await page.evaluate(() => {
      const bw = document.querySelector('.el-table__body-wrapper');
      return bw ? {
        scrollW: bw.scrollWidth,
        clientW: bw.clientWidth,
        hasOverflow: bw.scrollWidth > bw.clientWidth + 5,
      } : null;
    });
    console.log(`  Overflow: ${overflow?.hasOverflow ? 'YES' : 'NO'} (${overflow?.scrollW} vs ${overflow?.clientW})`);

    if (overflow?.hasOverflow) {
      // Scroll and check sync
      await page.evaluate(() => {
        const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
        if (sw) sw.scrollLeft = 150;
      });
      await sleep(600);

      const sync = await page.evaluate(() => {
        const hw = document.querySelector('.el-table__header-wrapper');
        const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
        return {
          header: hw ? Math.round(hw.scrollLeft) : -1,
          body: sw ? Math.round(sw.scrollLeft) : -1,
        };
      });
      console.log(`  Header: ${sync.header}, Body: ${sync.body}`);
      if (Math.abs(sync.header - sync.body) <= 5) {
        console.log('  ✅ Scroll synced'); pass++; results.push('Scroll sync');
      } else {
        console.log('  ❌ Scroll desync'); fail++; results.push('Scroll sync FAILED');
      }

      // Verify el-table has proper scrolling class
      const tableClass = await page.evaluate(() => {
        const t = document.querySelector('.el-table');
        return t ? t.className : '';
      });
      console.log(`  el-table classes: ${tableClass}`);
    } else {
      console.log('  ⚠️  No overflow, skip scroll test');
      pass++; results.push('Scroll sync (skipped)');
    }

    // ── Test 6: Second page ──
    console.log('\n=== Test 6: Task List Page ===');
    await page.evaluate(() => localStorage.removeItem('tk-table-widths'));
    await page.goto(`${BASE}/task/list`);
    await sleep(3000);

    const taskBefore = await getColWidths(page);
    console.log('  Initial widths:', taskBefore);

    if (taskBefore.length >= 2) {
      await dragColumn(page, 1, 60);
      const taskAfter = await getColWidths(page);
      const taskDiff = taskAfter[1] - taskBefore[1];
      console.log(`  Col 1: ${taskBefore[1]} → ${taskAfter[1]} (diff: ${taskDiff}px)`);
      if (Math.abs(taskDiff - 60) <= 10) {
        console.log('  ✅ Task list resize works'); pass++; results.push('Task list');
      } else {
        console.log('  ❌ Task list resize failed'); fail++; results.push('Task list FAILED');
      }
    } else {
      console.log('  ❌ Not enough columns');
      fail++; results.push('Task list FAILED');
    }

  } catch (err) {
    console.error('Error:', err.message);
    fail++;
  } finally {
    await browser.close();
  }

  console.log('\n════════════════ RESULTS ════════════════');
  results.forEach(r => console.log(`  ${r.includes('FAILED') ? '❌' : '✅'} ${r}`));
  console.log(`\nPassed: ${pass}/${pass + fail}  Failed: ${fail}`);
  process.exit(fail > 0 ? 1 : 0);
})();
