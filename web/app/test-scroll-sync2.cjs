/**
 * Correct scroll sync test: scrolls the el-scrollbar__wrap (not body-wrapper)
 * and verifies header scrolls in sync.
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

  await page.goto(`${BASE}/login`);
  await sleep(2000);
  await page.evaluate(() => {
    localStorage.setItem('tk-token', JSON.stringify('test-token'));
    localStorage.setItem('tk-table-widths', JSON.stringify({
      'asset-list': { name: 250, assetType: 200, assetKey: 300, status: 150, labels: 250, lastActiveAt: 200 }
    }));
  });
  await page.goto(`${BASE}/asset/list`);
  await sleep(4000);

  // Check widths
  const widths = await page.evaluate(() => {
    const cols = document.querySelectorAll('.el-table__header colgroup col');
    return Array.from(cols).map((c) => parseInt(c.getAttribute('width') || '0', 10));
  });
  console.log('Column widths:', widths);
  console.log('Total:', widths.reduce((a, b) => a + b, 0));

  // Check body scrollbar wrap - the actual scroll container
  const scrollbarInfo = await page.evaluate(() => {
    const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
    const hw = document.querySelector('.el-table__header-wrapper');
    if (!sw || !hw) return null;
    return {
      scrollbarWrapScrollW: sw.scrollWidth,
      scrollbarWrapClientW: sw.clientWidth,
      scrollbarWrapOverflow: getComputedStyle(sw).overflow,
      scrollbarWrapScrollable: sw.scrollWidth > sw.clientWidth + 5,
      headerScrollW: hw.scrollWidth,
      headerClientW: hw.clientWidth,
      headerScrollable: hw.scrollWidth > hw.clientWidth + 5,
    };
  });
  console.log('\nScrollbar info:', JSON.stringify(scrollbarInfo, null, 2));

  if (scrollbarInfo?.scrollbarWrapScrollable) {
    console.log('\n✅ Body scrollbar wrap is scrollable!');

    // Scroll the body horizontally
    await page.evaluate(() => {
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      if (sw) sw.scrollLeft = 200;
    });
    await sleep(800);

    const sync = await page.evaluate(() => {
      const hw = document.querySelector('.el-table__header-wrapper');
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      return {
        headerScrollLeft: hw ? Math.round(hw.scrollLeft) : -1,
        bodyScrollLeft: sw ? Math.round(sw.scrollLeft) : -1,
        diff: hw && sw ? Math.abs(hw.scrollLeft - sw.scrollLeft) : -1,
      };
    });
    console.log('After scroll (200px):', JSON.stringify(sync, null, 2));

    if (sync.diff <= 5) {
      console.log('✅ Header/Body scroll SYNC perfectly!');
    } else {
      console.log('❌ Scroll DESYNC! Header:', sync.headerScrollLeft, 'Body:', sync.bodyScrollLeft);
    }

    // Test scrolling to end
    await page.evaluate(() => {
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      if (sw) sw.scrollLeft = sw.scrollWidth;
    });
    await sleep(800);

    const syncEnd = await page.evaluate(() => {
      const hw = document.querySelector('.el-table__header-wrapper');
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      return {
        headerScrollLeft: hw ? Math.round(hw.scrollLeft) : -1,
        bodyScrollLeft: sw ? Math.round(sw.scrollLeft) : -1,
        diff: hw && sw ? Math.abs(hw.scrollLeft - sw.scrollLeft) : -1,
      };
    });
    console.log('After scroll to end:', JSON.stringify(syncEnd, null, 2));

    if (syncEnd.diff <= 5) {
      console.log('✅ Header/Body end-scroll SYNC perfectly!');
    } else {
      console.log('❌ Scroll DESYNC at end!');
    }

    // Scroll back to start
    await page.evaluate(() => {
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      if (sw) sw.scrollLeft = 0;
    });
    await sleep(800);

    const syncStart = await page.evaluate(() => {
      const hw = document.querySelector('.el-table__header-wrapper');
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      return {
        headerScrollLeft: hw ? Math.round(hw.scrollLeft) : -1,
        bodyScrollLeft: sw ? Math.round(sw.scrollLeft) : -1,
      };
    });
    console.log('After scroll to start:', JSON.stringify(syncStart, null, 2));

  } else {
    console.log('❌ Body scrollbar wrap is NOT scrollable');
  }

  // Also check: drag a column wider and verify scroll appears
  console.log('\n=== Drag to create overflow ===');
  await page.evaluate(() => localStorage.removeItem('tk-table-widths'));
  await page.goto(`${BASE}/asset/list`);
  await sleep(3000);

  // Drag column 1 wider
  const cellInfo = await page.evaluate(() => {
    const cells = document.querySelectorAll('.el-table__header-wrapper th.el-table__cell');
    if (cells.length < 2) return null;
    const cell = cells[1];
    const rect = cell.getBoundingClientRect();
    return { right: rect.right, top: rect.top, bottom: rect.bottom };
  });

  if (cellInfo) {
    const startX = cellInfo.right - 3;
    const startY = (cellInfo.top + cellInfo.bottom) / 2;
    
    await page.mouse.move(startX, startY);
    await sleep(50);
    await page.mouse.down();
    await sleep(80);
    
    // Drag +300px to force overflow
    for (let i = 1; i <= 20; i++) {
      await page.mouse.move(startX + (300 * i) / 20, startY);
      await sleep(15);
    }
    await page.mouse.up();
    await sleep(800);

    // Check if scrollbar appeared after drag
    const afterDragScroll = await page.evaluate(() => {
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      if (!sw) return null;
      return {
        scrollW: sw.scrollWidth,
        clientW: sw.clientWidth,
        scrollable: sw.scrollWidth > sw.clientWidth + 5,
      };
    });
    console.log('After +300px drag:', JSON.stringify(afterDragScroll, null, 2));

    if (afterDragScroll?.scrollable) {
      console.log('✅ Horizontal scroll appeared after dragging column wider!');
      
      // Test scroll sync after drag
      await page.evaluate(() => {
        const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
        if (sw) sw.scrollLeft = 150;
      });
      await sleep(600);

      const syncAfterDrag = await page.evaluate(() => {
        const hw = document.querySelector('.el-table__header-wrapper');
        const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
        return {
          header: hw ? Math.round(hw.scrollLeft) : -1,
          body: sw ? Math.round(sw.scrollLeft) : -1,
          diff: hw && sw ? Math.abs(hw.scrollLeft - sw.scrollLeft) : -1,
        };
      });
      console.log('Scroll sync after drag:', JSON.stringify(syncAfterDrag, null, 2));
      if (syncAfterDrag.diff <= 5) {
        console.log('✅ Perfect scroll sync after drag!');
      }
    }
  }

  await browser.close();
})();
