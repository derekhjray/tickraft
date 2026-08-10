/**
 * Test scroll sync using widths from the previous test run.
 * Uses the already-persisted widths to force overflow.
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

  // Set wide widths BEFORE loading page (using route interception to inject)
  // Actually, let's just set it via page.evaluate then navigate
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
  const total = widths.reduce((a, b) => a + b, 0);
  console.log('Total:', total);

  // Check overflow
  const overflow = await page.evaluate(() => {
    const bw = document.querySelector('.el-table__body-wrapper');
    const hw = document.querySelector('.el-table__header-wrapper');
    if (!bw || !hw) return null;
    return {
      bodyScrollW: bw.scrollWidth,
      bodyClientW: bw.clientWidth,
      bodyOverflow: bw.scrollWidth > bw.clientWidth + 5,
      headerScrollW: hw.scrollWidth,
      headerClientW: hw.clientWidth,
      headerOverflow: hw.scrollWidth > hw.clientWidth + 5,
    };
  });
  console.log('Overflow:', JSON.stringify(overflow, null, 2));

  if (overflow?.bodyOverflow && overflow?.headerOverflow) {
    console.log('\n✅ Both header and body have overflow — testing scroll sync...');
    
    // Scroll body wrapper horizontally
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
    console.log('Scroll sync:', JSON.stringify(sync, null, 2));

    if (sync.diff <= 5) {
      console.log('✅ Scroll sync works perfectly!');
    } else {
      console.log('❌ Scroll desync detected!');
    }
  } else {
    console.log('\n⚠️  No overflow detected. Checking if widths were applied...');
    // Check what widths are actually in localStorage
    const stored = await page.evaluate(() => localStorage.getItem('tk-table-widths'));
    console.log('Stored widths:', stored);
    
    // Try reading them from the console
    const debugInfo = await page.evaluate(() => {
      const cols = document.querySelectorAll('.el-table__header colgroup col');
      const result = [];
      for (let i = 0; i < cols.length; i++) {
        result.push({
          width: cols[i].getAttribute('width'),
          computedWidth: cols[i].getBoundingClientRect().width,
        });
      }
      return result;
    });
    console.log('Col details:', JSON.stringify(debugInfo, null, 2));
  }

  await browser.close();
})();
