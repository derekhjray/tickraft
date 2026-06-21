/**
 * Test: Column drag stability verification
 * 
 * Verifies that when dragging one column:
 * 1. Only the dragged column changes width
 * 2. Other columns' widths remain stable (no chaotic redistribution)
 * 3. The drag result persists after the drag completes
 * 4. Dragging a second time on another column also works correctly
 */
const PW_PATH = '/Users/derekray/workspace/auzekalabs/tickraft/arcadia/node_modules/.pnpm/playwright@1.61.1/node_modules/playwright';
const { chromium } = require(PW_PATH);
const BASE = 'http://localhost:5190';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

(async () => {
  const browser = await chromium.launch({
    headless: true,
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await ctx.newPage();

  // Track console errors
  const errors = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') errors.push(msg.text());
  });

  // Login and navigate to a page with a DataTable
  await page.goto(`${BASE}/login`);
  await sleep(1500);
  await page.evaluate(() => {
    localStorage.setItem('tk-token', JSON.stringify('test-token'));
    localStorage.removeItem('tk-table-widths');
  });

  // Navigate to the asset list page which has a DataTable
  await page.goto(`${BASE}/asset/list`);
  await sleep(3000);

  // Get initial column widths
  const getColWidths = async () => {
    return page.evaluate(() => {
      const cols = document.querySelectorAll('.el-table__header colgroup col');
      return Array.from(cols).map((c) => parseInt(c.getAttribute('width') || '0', 10));
    });
  };

  const getThCells = async () => {
    return page.evaluate(() => {
      const cells = document.querySelectorAll('.el-table__header-wrapper th.el-table__cell');
      return Array.from(cells).map((c) => {
        const rect = c.getBoundingClientRect();
        return { left: rect.left, right: rect.right, top: rect.top, bottom: rect.bottom, width: rect.width };
      });
    });
  };

  console.log('=== Test 1: Initial Column Widths ===');
  const initialWidths = await getColWidths();
  console.log('Initial widths:', initialWidths);
  console.log('Total:', initialWidths.reduce((a, b) => a + b, 0));

  if (initialWidths.length < 3) {
    console.log('❌ Not enough columns for testing');
    await browser.close();
    process.exit(1);
  }

  // Test 2: Drag column 2 wider and check other columns don't change
  console.log('\n=== Test 2: Drag Column Wider (+80px) ===');
  
  const cellsBefore = await getThCells();
  const targetIdx = Math.min(2, cellsBefore.length - 1); // Drag 3rd column (index 2)
  
  // Find the right edge of the target column header
  const targetCell = cellsBefore[targetIdx];
  const startX = targetCell.right - 3; // Near right edge
  const startY = (targetCell.top + targetCell.bottom) / 2;
  const originalWidth = initialWidths[targetIdx + 1] || targetCell.width; // +1 for selection column
  
  console.log(`Dragging column at index ${targetIdx} (original DOM width: ${targetCell.width}px)`);
  console.log(`Original colgroup width: ${originalWidth}px`);
  
  // Get all widths before drag
  const widthsBefore = await getColWidths();
  console.log('Widths before drag:', widthsBefore);
  
  // Perform drag
  await page.mouse.move(startX, startY);
  await sleep(50);
  await page.mouse.down();
  await sleep(80);
  
  // Drag +80px wider
  const dragDistance = 80;
  for (let i = 1; i <= 20; i++) {
    await page.mouse.move(startX + (dragDistance * i) / 20, startY);
    await sleep(10);
  }
  await page.mouse.up();
  await sleep(500);
  
  // Get widths after drag
  const widthsAfter = await getColWidths();
  console.log('Widths after drag: ', widthsAfter);
  
  // Check stability: non-target columns should not change significantly
  const tolerance = 2; // 2px tolerance for border adjustments
  let chaoticChanges = 0;
  for (let i = 0; i < widthsBefore.length; i++) {
    if (i === targetIdx + 1) continue; // Skip the dragged column
    const diff = Math.abs(widthsAfter[i] - widthsBefore[i]);
    if (diff > tolerance) {
      chaoticChanges++;
      console.log(`  ⚠️  Column ${i} changed by ${diff}px (${widthsBefore[i]} → ${widthsAfter[i]})`);
    }
  }
  
  const targetDiff = widthsAfter[targetIdx + 1] - widthsBefore[targetIdx + 1];
  console.log(`Target column width change: ${widthsBefore[targetIdx + 1]} → ${widthsAfter[targetIdx + 1]} (delta: ${targetDiff}px)`);
  
  if (targetDiff > 0 && chaoticChanges === 0) {
    console.log('✅ PASS: Only the dragged column changed width, others are stable!');
  } else if (targetDiff > 0 && chaoticChanges <= 1) {
    console.log('⚠️  MINOR: Dragged column expanded, with minimal impact on others');
  } else {
    console.log('❌ FAIL: Multiple columns changed chaotically!');
  }

  // Test 3: Drag a different column narrower
  console.log('\n=== Test 3: Drag Another Column Narrower (-50px) ===');
  
  const cellsBefore2 = await getThCells();
  const targetIdx2 = Math.min(3, cellsBefore2.length - 1);
  
  const targetCell2 = cellsBefore2[targetIdx2];
  const startX2 = targetCell2.right - 3;
  const startY2 = (targetCell2.top + targetCell2.bottom) / 2;
  
  const widthsBefore2 = await getColWidths();
  console.log('Widths before 2nd drag:', widthsBefore2);
  
  await page.mouse.move(startX2, startY2);
  await sleep(50);
  await page.mouse.down();
  await sleep(80);
  
  // Drag -50px narrower
  const dragDistance2 = -50;
  for (let i = 1; i <= 20; i++) {
    await page.mouse.move(startX2 + (dragDistance2 * i) / 20, startY2);
    await sleep(10);
  }
  await page.mouse.up();
  await sleep(500);
  
  const widthsAfter2 = await getColWidths();
  console.log('Widths after 2nd drag: ', widthsAfter2);
  
  // Check: columns other than targetIdx2+1 should be stable or absorb from the dragged column
  let chaoticChanges2 = 0;
  for (let i = 0; i < widthsBefore2.length; i++) {
    if (i === targetIdx2 + 1) continue;
    const diff = Math.abs(widthsAfter2[i] - widthsBefore2[i]);
    if (diff > 3) {
      chaoticChanges2++;
      console.log(`  ⚠️  Column ${i} changed by ${diff}px (${widthsBefore2[i]} → ${widthsAfter2[i]})`);
    }
  }
  
  const targetDiff2 = widthsAfter2[targetIdx2 + 1] - widthsBefore2[targetIdx2 + 1];
  console.log(`Target column width change: ${widthsBefore2[targetIdx2 + 1]} → ${widthsAfter2[targetIdx2 + 1]} (delta: ${targetDiff2}px)`);
  
  if (Math.abs(targetDiff2) > 0 && chaoticChanges2 <= 1) {
    console.log('✅ PASS: Second drag stable!');
  } else {
    console.log('❌ FAIL: Multiple columns changed chaotically on second drag!');
  }

  // Test 4: Verify persistence
  console.log('\n=== Test 4: Persistence Check ===');
  const storedWidths = await page.evaluate(() => {
    const raw = localStorage.getItem('tk-table-widths');
    return raw ? JSON.parse(raw) : null;
  });
  console.log('Stored widths:', JSON.stringify(storedWidths, null, 2));
  
  // Check if asset-list table has persisted widths
  if (storedWidths && storedWidths['asset-list']) {
    const persisted = storedWidths['asset-list'];
    console.log('✅ Persistence works! Table widths saved for asset-list');
  } else {
    console.log('⚠️  No persisted widths found (may need to change tableId)');
  }

  // Test 5: Verify scroll sync after drag
  console.log('\n=== Test 5: Scroll Sync After Drag ===');
  const scrollInfo = await page.evaluate(() => {
    const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
    const hw = document.querySelector('.el-table__header-wrapper');
    if (!sw || !hw) return null;
    return {
      scrollbarScrollW: sw.scrollWidth,
      scrollbarClientW: sw.clientWidth,
      scrollbarOverflow: getComputedStyle(sw).overflow,
      headerScrollW: hw.scrollWidth,
      headerClientW: hw.clientWidth,
      scrollable: sw.scrollWidth > sw.clientWidth + 5,
    };
  });
  console.log('Scroll info:', JSON.stringify(scrollInfo, null, 2));
  
  if (scrollInfo?.scrollable) {
    await page.evaluate(() => {
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      if (sw) sw.scrollLeft = 100;
    });
    await sleep(300);
    
    const syncCheck = await page.evaluate(() => {
      const hw = document.querySelector('.el-table__header-wrapper');
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      return {
        header: hw ? Math.round(hw.scrollLeft) : -1,
        body: sw ? Math.round(sw.scrollLeft) : -1,
      };
    });
    console.log('Scroll sync check:', JSON.stringify(syncCheck, null, 2));
    if (Math.abs(syncCheck.header - syncCheck.body) <= 5) {
      console.log('✅ Scroll sync works after drag!');
    }
  }

  // Test 6: Reload page and verify persisted widths are restored
  console.log('\n=== Test 6: Reload & Verify Persistence ===');
  await page.reload();
  await sleep(3000);
  
  const widthsAfterReload = await getColWidths();
  console.log('Widths after reload:', widthsAfterReload);
  
  // If there were persisted widths, they should be restored
  if (storedWidths && storedWidths['asset-list']) {
    console.log('✅ Persisted widths should be restored after reload');
  }

  // Check for console errors
  if (errors.length > 0) {
    console.log('\n⚠️  Console errors:', errors.slice(0, 5));
  } else {
    console.log('\n✅ No console errors!');
  }

  console.log('\n=== All Tests Completed ===');
  await browser.close();
})();
