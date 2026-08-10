/**
 * Deep diagnostic: analyze why header overflows but body doesn't.
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

  const diag = await page.evaluate(() => {
    const result = {};

    // Table element
    const table = document.querySelector('.el-table');
    result.tableClass = table?.className;
    result.tableStyle = table?.getAttribute('style');

    // Header wrapper
    const hw = document.querySelector('.el-table__header-wrapper');
    result.headerWrapper = {
      clientW: hw?.clientWidth,
      scrollW: hw?.scrollWidth,
      style: hw?.getAttribute('style'),
      computedOverflow: hw ? getComputedStyle(hw).overflow : 'n/a',
    };

    // Header table
    const headerTable = document.querySelector('.el-table__header-wrapper table');
    result.headerTable = {
      offsetWidth: headerTable?.offsetWidth,
      style: headerTable?.getAttribute('style'),
      tableLayout: headerTable ? getComputedStyle(headerTable).tableLayout : 'n/a',
    };

    // Body wrapper  
    const bw = document.querySelector('.el-table__body-wrapper');
    result.bodyWrapper = {
      clientW: bw?.clientWidth,
      scrollW: bw?.scrollWidth,
      style: bw?.getAttribute('style'),
      computedOverflow: bw ? getComputedStyle(bw).overflow : 'n/a',
      computedOverflowX: bw ? getComputedStyle(bw).overflowX : 'n/a',
    };

    // Body scrollbar wrap
    const bsw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
    result.bodyScrollbarWrap = {
      clientW: bsw?.clientWidth,
      scrollW: bsw?.scrollWidth,
      style: bsw?.getAttribute('style'),
      computedOverflow: bsw ? getComputedStyle(bsw).overflow : 'n/a',
    };

    // Body table
    const bodyTable = document.querySelector('.el-table__body-wrapper table');
    result.bodyTable = {
      offsetWidth: bodyTable?.offsetWidth,
      style: bodyTable?.getAttribute('style'),
      tableLayout: bodyTable ? getComputedStyle(bodyTable).tableLayout : 'n/a',
    };

    // Body cells (first row)
    const bodyCells = document.querySelectorAll('.el-table__body-wrapper tbody tr:first-child td');
    result.bodyCells = Array.from(bodyCells).map((td, i) => ({
      index: i,
      width: td.getBoundingClientRect().width,
      computedStyle: getComputedStyle(td).width,
    }));

    // Header cells (first row)
    const headerCells = document.querySelectorAll('.el-table__header-wrapper thead tr:first-child th');
    result.headerCells = Array.from(headerCells).map((th, i) => ({
      index: i,
      width: th.getBoundingClientRect().width,
      computedStyle: getComputedStyle(th).width,
    }));

    return result;
  });

  console.log('=== Deep Diagnostics ===');
  console.log('\nTable class:', diag.tableClass);
  console.log('Table style:', diag.tableStyle);

  console.log('\nHeader wrapper:', JSON.stringify(diag.headerWrapper, null, 2));
  console.log('Header table:', JSON.stringify(diag.headerTable, null, 2));

  console.log('\nBody wrapper:', JSON.stringify(diag.bodyWrapper, null, 2));
  console.log('Body scrollbar wrap:', JSON.stringify(diag.bodyScrollbarWrap, null, 2));
  console.log('Body table:', JSON.stringify(diag.bodyTable, null, 2));

  console.log('\nHeader cells:');
  diag.headerCells.forEach(c => console.log(`  Col ${c.index}: width=${c.width}, computed=${c.computedStyle}`));

  console.log('\nBody cells:');
  diag.bodyCells.forEach(c => console.log(`  Col ${c.index}: width=${c.width}, computed=${c.computedStyle}`));

  // Test scroll sync if both overflow
  const bothOverflow = diag.headerWrapper.scrollW > diag.headerWrapper.clientW + 5
    && diag.bodyWrapper.scrollW > diag.bodyWrapper.clientW + 5;
  console.log('\nBoth overflow:', bothOverflow);

  // Try forcing scroll on body wrapper
  console.log('\n=== Forcing scroll sync test ===');
  await page.evaluate(() => {
    // Check if body wrapper's table needs width override
    const bodyTable = document.querySelector('.el-table__body-wrapper table');
    if (bodyTable) {
      console.log('Body table current width:', bodyTable.style.width || 'auto');
      console.log('Body table offsetWidth:', bodyTable.offsetWidth);
      
      // Try setting explicit width
      const totalColWidth = Array.from(document.querySelectorAll('.el-table__header colgroup col'))
        .reduce((sum, c) => sum + parseInt(c.getAttribute('width') || '0', 10), 0);
      console.log('Total column width:', totalColWidth);
      
      if (bodyTable.offsetWidth < totalColWidth) {
        console.log('⚠️ Body table is narrower than total column width!');
        console.log('Setting body table width to', totalColWidth, 'px...');
        bodyTable.style.width = totalColWidth + 'px';
      }
    }
  });

  const afterFix = await page.evaluate(() => {
    const bw = document.querySelector('.el-table__body-wrapper');
    return {
      bodyScrollW: bw?.scrollWidth,
      bodyClientW: bw?.clientWidth,
      bodyOverflow: bw ? bw.scrollWidth > bw.clientWidth + 5 : false,
    };
  });
  console.log('After fix:', JSON.stringify(afterFix, null, 2));

  // Try scrolling now
  if (afterFix.bodyOverflow) {
    await page.evaluate(() => {
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      if (sw) sw.scrollLeft = 200;
    });
    await sleep(500);
    const syncAfterFix = await page.evaluate(() => {
      const hw = document.querySelector('.el-table__header-wrapper');
      const sw = document.querySelector('.el-table__body-wrapper .el-scrollbar__wrap');
      return {
        header: hw ? Math.round(hw.scrollLeft) : -1,
        body: sw ? Math.round(sw.scrollLeft) : -1,
        diff: hw && sw ? Math.abs(hw.scrollLeft - sw.scrollLeft) : -1,
      };
    });
    console.log('Scroll sync after fix:', JSON.stringify(syncAfterFix, null, 2));
  }

  await browser.close();
})();
